/*
Package watcher provides an inotify-based approach for watching file system events from a directory recursively.
*/
package watcher

import (
	"fmt"
	"io/ioutil"
	"path"
	"regexp"
	"strings"
	"unsafe"

	"golang.org/x/sys/unix"
)

const eventsBufferSize = (unix.SizeofInotifyEvent + 1 + unix.NAME_MAX) * 64
const inotifyMask = unix.IN_CREATE | unix.IN_DELETE | unix.IN_CLOSE_WRITE | unix.IN_MOVED_FROM | unix.IN_MOVED_TO

// W is a watcher for a directory.
type W struct {
	fd            int
	closed        bool
	tree          *watchedDirsTree
	ignoreRegExps []*regexp.Regexp
	done          chan struct{}
	events        chan Event
	errs          chan error
	mvEvents      *mvEvents
}

// New creates a watcher for dirPath recursively, ignoring any path that matches at least one of ignoreRegExps.
// Directory paths matched against ignoreRegExps end with a /.
func New(dirPath string, ignoreRegExps []*regexp.Regexp) (*W, error) {
	fd, err := unix.InotifyInit1(0)
	if err != nil {
		return nil, fmt.Errorf("creating inotify instance: %v", err)
	}

	done := make(chan struct{})
	w := &W{
		fd:            fd,
		tree:          newWatchedDirsTree(),
		done:          done,
		ignoreRegExps: ignoreRegExps,
	}

	rootWd, err := w.addToInotify(dirPath)
	if err != nil {
		return nil, err
	}
	w.tree.setRoot(dirPath, rootWd)

	err = w.addDirsStartingAt(dirPath)
	if err != nil {
		return nil, err
	}

	w.events = make(chan Event)
	w.errs = make(chan error)
	w.mvEvents = newMvEvents()

	w.startReading()

	return w, nil
}

func (w *W) startReading() {
	readingErr := make(chan error)
	readingRes := make(chan struct {
		inotifyE unix.InotifyEvent
		name     string
	})

	// reading from inotify instance's fd
	go func() {
		buff := [eventsBufferSize]byte{}

		for {
			select {
			case <-w.done:
				return
			default:
			}

			n, err := unix.Read(w.fd, buff[:])
			if err != nil {
				readingErr <- err
				return
			}

			previousNameLen := 0
			for i := 0; i < n; i += int(unix.SizeofInotifyEvent + previousNameLen) {
				select {
				case <-w.done:
					return
				default:
				}

				var name string

				inotifyE := (*unix.InotifyEvent)(unsafe.Pointer(&buff[i]))

				if inotifyE.Len > 0 {
					name = string(buff[i+unix.SizeofInotifyEvent : i+int(unix.SizeofInotifyEvent+inotifyE.Len)])
					// remove trailing null chars
					name = strings.TrimRight(name, "\x00")
				}

				readingRes <- struct {
					inotifyE unix.InotifyEvent
					name     string
				}{
					*inotifyE,
					name,
				}

				previousNameLen = int(inotifyE.Len)
			}
		}
	}()

	go func() {
		defer w.mvEvents.close()
		defer w.Close()

		for {
			select {
			case <-w.done:
				return
			case err := <-readingErr:
				w.errs <- fmt.Errorf("reading from inotify instance's fd: %v", err)

				return
			case res := <-readingRes:
				var e Event

				parentDir := w.tree.get(int(res.inotifyE.Wd))
				// this happens when an IN_IGNORED event about an already
				// removed directory is received.
				if parentDir == nil {
					continue
				}

				isDir := res.inotifyE.Mask&unix.IN_ISDIR == unix.IN_ISDIR

				fileOrDirPath := path.Join(w.tree.path(parentDir.wd), res.name)
				// if it matches, it means it should be ignored
				if w.matchPath(fileOrDirPath, isDir) {
					continue
				}

				switch {
				// this event is only handled if it is from the root,
				// since, if it is from any other directory, it means
				// that this directory's parent has already received
				// an IN_DELETE event and the directory's been already
				// removed from the inotify instance and the tree.
				case res.inotifyE.Mask&unix.IN_IGNORED == unix.IN_IGNORED && w.tree.get(int(res.inotifyE.Wd)) == w.tree.root:
					return
				case res.inotifyE.Mask&unix.IN_CREATE == unix.IN_CREATE:
					if isDir {
						_, match, err := w.addDir(res.name, parentDir.wd)
						if !match {
							if err != nil {
								w.errs <- err

								return
							}

							err = w.addDirsStartingAt(fileOrDirPath)
							if err != nil {
								w.errs <- err

								return
							}
						}
					}

					e = CreateEvent{
						path:  fileOrDirPath,
						isDir: isDir,
					}
				case res.inotifyE.Mask&unix.IN_DELETE == unix.IN_DELETE:
					if isDir {
						dir := w.tree.find(fileOrDirPath)
						// this should never happen
						if dir == nil {
							continue
						}

						// the directory isn't removed from the inotify instance
						// because it was removed automatically when it was removed
						w.tree.rm(dir.wd)
					}

					e = DeleteEvent{
						path:  fileOrDirPath,
						isDir: isDir,
					}
				case res.inotifyE.Mask&unix.IN_CLOSE_WRITE == unix.IN_CLOSE_WRITE:
					e = ModifyEvent{
						path: fileOrDirPath,
					}
				case res.inotifyE.Mask&unix.IN_MOVED_FROM == unix.IN_MOVED_FROM:
					w.mvEvents.addMvFrom(int(res.inotifyE.Cookie), res.name, int(res.inotifyE.Wd), isDir)
				case res.inotifyE.Mask&unix.IN_MOVED_TO == unix.IN_MOVED_TO:
					w.mvEvents.addMvTo(int(res.inotifyE.Cookie), res.name, int(res.inotifyE.Wd), isDir)
				}

				if e != nil {
					w.events <- e
				}
			case mvEvent := <-w.mvEvents.queue:
				var oldPath, newPath string

				hasMvFrom := mvEvent.oldName != ""
				hasMvTo := mvEvent.newName != ""

				switch {
				case hasMvFrom && hasMvTo:
					oldPath = path.Join(
						w.tree.path(mvEvent.oldParentWd),
						mvEvent.oldName,
					)
					newPath = path.Join(
						w.tree.path(mvEvent.newParentWd),
						mvEvent.newName,
					)

					if mvEvent.isDir {
						w.tree.mv(w.tree.find(oldPath).wd, mvEvent.newParentWd, mvEvent.newName)
					}
				case hasMvFrom:
					oldPath = path.Join(
						w.tree.path(mvEvent.oldParentWd),
						mvEvent.oldName,
					)

					if mvEvent.isDir {
						w.tree.rm(w.tree.find(oldPath).wd)
					}
				case hasMvTo:
					newPath = path.Join(
						w.tree.path(mvEvent.newParentWd),
						mvEvent.newName,
					)

					if mvEvent.isDir {
						_, match, err := w.addDir(mvEvent.newName, mvEvent.newParentWd)
						if !match {
							if err != nil {
								w.errs <- err

								return
							}

							err = w.addDirsStartingAt(newPath)
							if err != nil {
								w.errs <- err

								return
							}
						}
					}
				}

				w.events <- RenameEvent{
					isDir:   mvEvent.isDir,
					OldPath: oldPath,
					path:    newPath,
				}
			}
		}
	}()
}

// Events returns the events channel.
func (w *W) Events() chan Event {
	return w.events
}

// Errs returns the errors channel.
func (w *W) Errs() chan error {
	return w.errs
}

// Wait blocks until the watcher is closed.
func (w *W) Wait() {
	<-w.done
}

// Close closes the watcher.
// If the watcher is already closed, it's a no-op.
func (w *W) Close() error {
	if w.closed {
		return nil
	}

	w.closed = true
	err := unix.Close(w.fd)
	close(w.done)
	if err != nil {
		return fmt.Errorf("closing fd: %v", err)
	}

	return nil
}

// addDirsStartingAt adds every directory descendant of rootPath recursively
// to the tree and to the inotify instance.
// This functions assumes that there's a node in the tree whose path is equal
// to cleanPath(rootPath).
func (w *W) addDirsStartingAt(rootPath string) error {
	entries, err := ioutil.ReadDir(rootPath)
	if err != nil {
		return fmt.Errorf("reading %v dir: %v", rootPath, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			dirPath := path.Join(rootPath, entry.Name())

			_, match, err := w.addDir(
				entry.Name(),
				w.tree.find(cleanPath(rootPath)).wd,
			)
			if match {
				continue
			}
			if err != nil {
				return err
			}

			err = w.addDirsStartingAt(dirPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// addDir checks if a directory isn't a match for any of w.ignoreRegExps and, if it isn't,
// adds it to the tree and to the inotify instance and returns the added directory's wd.
func (w *W) addDir(name string, parentWd int) (wd int, match bool, err error) {
	dirPath := path.Join(w.tree.path(parentWd), name)

	if w.matchPath(dirPath, true) {
		return -1, true, nil
	}

	wd, err = w.addToInotify(dirPath)
	if err != nil {
		return -1, false, err
	}

	w.tree.add(wd, name, parentWd)

	return wd, false, nil
}

// addToInotify adds the given path to the inotify instance and returns the added
// directory's wd.
// Note that it doesn't check whether the given path is match for any of
// w.ignoreRegExps.
func (w *W) addToInotify(path string) (int, error) {
	wd, err := unix.InotifyAddWatch(w.fd, path, inotifyMask)
	if err != nil {
		return -1, fmt.Errorf("adding directory to inotify instance: %v", err)
	}

	return wd, nil
}

// removeFromInotify removes the given path from the inotify instance.
func (w *W) removeFromInotify(wd int) error {
	wd, err := unix.InotifyRmWatch(w.fd, uint32(wd))
	if err != nil {
		return fmt.Errorf("removing directory from inotify instance: %v", err)
	}

	return nil
}

// matchPath returns whether the given path matchs any of w.ignoreRegExps.
func (w *W) matchPath(path string, isDir bool) bool {
	if isDir {
		path += "/"
	}

	for _, rx := range w.ignoreRegExps {
		if match := rx.MatchString(path); match {
			return true
		}
	}

	return false
}
