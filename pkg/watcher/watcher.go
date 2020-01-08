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
const inotifyMask = unix.IN_CLOSE_WRITE | unix.IN_CREATE | unix.IN_MOVED_FROM | unix.IN_MOVED_TO

// W is a watcher for the working directory.
type W struct {
	fd            int
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
					w.Close()
					return
				}

				n = res.n
			}

		buffLoop:
			for i := 0; i < n; {
				var name string
				inotifyE := (*unix.InotifyEvent)(unsafe.Pointer(&buff[i]))

				if inotifyE.Len > 0 {
					name = string(buff[i+unix.SizeofInotifyEvent : i+int(unix.SizeofInotifyEvent+inotifyE.Len)])
					name = strings.TrimRight(name, "\x00")
				}

				dir := w.tree.get(int(inotifyE.Wd))
				var e Event

				fileOrDirPath := path.Join(w.tree.path(dir.wd), name)
				isDir := inotifyE.Mask&unix.IN_ISDIR == unix.IN_ISDIR

				switch {
				case inotifyE.Mask&unix.IN_CREATE == unix.IN_CREATE:
					if isDir {
						_, match, err := w.addDir(name, dir.wd)
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
					} else {
						if w.matchPath(fileOrDirPath) {
							continue buffLoop
						}
					}

					e = CreateEvent{
						path:  fileOrDirPath,
						isDir: isDir,
					}
				}

				events <- e

				i += int(unix.SizeofInotifyEvent + inotifyE.Len)
			}
		}
	}()

	return events, errs
}

// Close closes the watcher.
func (w *W) Close() error {
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

// addDir checks if a directory isn't a match for any of w.ignoreRegExps and, if it isn't, adds it
// to the tree and to the inotify instance and returns the added directory's wd.
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
func (w *W) addToInotify(path string) (int, error) {
	wd, err := unix.InotifyAddWatch(w.fd, path, inotifyMask)
	if err != nil {
		return -1, fmt.Errorf("adding directory to inotify instance: %v", err)
	}

	return wd, nil
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
