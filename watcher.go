package main

import (
	"os"

	"github.com/fsnotify/fsnotify"
)

func watch() {
	dir, _ := os.Getwd()
	watcher, _ := fsnotify.NewWatcher()

	watcher.Add(dir)

	for {
		e := <-watcher.Events

		logEvt.Print(e.String())
	}
}
