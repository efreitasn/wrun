/*
Package watcher provides an inotify-based approach for watching file system events on the working
directory and its descendants recursively.
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

// W is a watcher for the working directory.
type W struct {
	fd            int
	closed        bool
	tree          *watchedDirsTree
	ignoreRegExps []*regexp.Regexp
	done          chan struct{}
	Done          <-chan struct{}
}

// New creates a watcher for the working directory.
func New(ignoreRegExps []*regexp.Regexp) (*W, error) {
	fd, err := unix.InotifyInit1(0)
	if err != nil {
		return nil, fmt.Errorf("creating inotify instance: %v", err)
	}

	done := make(chan struct{})
	w := &W{
		fd:            fd,
		tree:          newWatchedDirsTree(),
		done:          done,
		Done:          done,
		ignoreRegExps: ignoreRegExps,
	}

	rootWd, err := w.addToInotify(".")
	if err != nil {
		return nil, err
	}
	w.tree.setRoot(rootWd)

	err = w.addDirsStartingAt(".")
	if err != nil {
		return nil, err
	}

	return w, nil
}

// Start starts the watcher.
func (w *W) Start() (events chan Event, errs chan error) {
	events = make(chan Event)
	errs = make(chan error)

	go func() {
		defer w.Close()

		buff := [eventsBufferSize]byte{}

		for {
			readRes := make(chan struct {
				n   int
				err error
			})
			go func() {
				n, err := unix.Read(w.fd, buff[:])
				readRes <- struct {
					n   int
					err error
				}{n, err}
			}()

			var n int

			select {
			case <-w.done:
				return
			case res := <-readRes:
				if res.err != nil {
					errs <- res.err
					return
				}

				n = res.n
			}

			var inotifyE *unix.InotifyEvent
		buffLoop:
			for i := 0; i < n; i += int(unix.SizeofInotifyEvent + inotifyE.Len) {
				select {
				case <-w.done:
					return
				default:
				}

				var name string
				inotifyE = (*unix.InotifyEvent)(unsafe.Pointer(&buff[i]))

				if inotifyE.Len > 0 {
					name = string(buff[i+unix.SizeofInotifyEvent : i+int(unix.SizeofInotifyEvent+inotifyE.Len)])
					// remove trailing null chars
					name = strings.TrimRight(name, "\x00")
				}

				parentDir := w.tree.get(int(inotifyE.Wd))
				var e Event

				fileOrDirPath := path.Join(w.tree.path(parentDir.wd), name)
				// if it matches, it means it should be ignored
				if w.matchPath(fileOrDirPath) {
					continue buffLoop
				}

				isDir := inotifyE.Mask&unix.IN_ISDIR == unix.IN_ISDIR

				switch {
				// this event is only handled if it is from the root,
				// since, if it is from any other directory, it means
				// that this directory's parent has already received
				// an IN_DELETE event and the directory's been already
				// removed from the inotify instance and the tree.
				case inotifyE.Mask&unix.IN_IGNORED == unix.IN_IGNORED && w.tree.get(int(inotifyE.Wd)) == w.tree.root:
					return
				case inotifyE.Mask&unix.IN_CREATE == unix.IN_CREATE:
					if isDir {
						_, match, err := w.addDir(name, parentDir.wd)
						if !match {
							if err != nil {
								errs <- err

								return
							}

							err = w.addDirsStartingAt(fileOrDirPath)
							if err != nil {
								errs <- err

								return
							}
						}
					}

					e = CreateEvent{
						path:  fileOrDirPath,
						isDir: isDir,
					}
				case inotifyE.Mask&unix.IN_DELETE == unix.IN_DELETE:
					if isDir {
						dir := w.tree.find(fileOrDirPath)
						// this should never happen
						if dir == nil {
							continue buffLoop
						}

						// the directory isn't removed from the inotify instance
						// because it was removed automatically when it was removed
						w.tree.rm(dir.wd)
					}

					e = DeleteEvent{
						path:  fileOrDirPath,
						isDir: isDir,
					}
				case inotifyE.Mask&unix.IN_CLOSE_WRITE == unix.IN_CLOSE_WRITE:
					e = ModifyEvent{
						path:  fileOrDirPath,
						isDir: isDir,
					}
				}

				if e != nil {
					events <- e
				}
			}
		}
	}()

	return events, errs
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

	return fmt.Errorf("closing fd: %v", err)
}

// addDirsStartingAt adds every directory descendant of rootPath recursively
// to the tree and to the inotify instance.
// This functions assumes that there's a node in the tree whose path is equal
// to path.Clean(rootPath).
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
				w.tree.find(path.Clean(rootPath)).wd,
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

	if w.matchPath(dirPath) {
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
	fmt.Println(wd, w.tree.get(wd))
	wd, err := unix.InotifyRmWatch(w.fd, uint32(wd))
	if err != nil {
		return fmt.Errorf("removing directory from inotify instance: %v", err)
	}

	return nil
}

// matchPath returns whether the given path matchs any of w.ignoreRegExps.
func (w *W) matchPath(path string) bool {
	for _, rx := range w.ignoreRegExps {
		if match := rx.MatchString(path); match {
			return true
		}
	}

	return false
}
