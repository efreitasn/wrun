package main

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/fsnotify/fsnotify"
)

func watchDir(dir string, watcher *fsnotify.Watcher) error {
	watcher.Add(dir)
	entries, err := ioutil.ReadDir(dir)

	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			err := watchDir(path.Join(dir, entry.Name()), watcher)

			if err != nil {
				return err
			}
		}
	}

	return nil
}

func createWatcher() (*fsnotify.Watcher, error) {
	dir, err := os.Getwd()

	if err != nil {
		return nil, err
	}

	watcher, err := fsnotify.NewWatcher()

	if err != nil {
		return nil, err
	}

	err = watchDir(dir, watcher)

	if err != nil {
		return nil, err
	}

	return watcher, nil
}
