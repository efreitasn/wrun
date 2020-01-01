package main

import (
	watcherLib "github.com/radovskyb/watcher"
)

func createWatcher() (*watcherLib.Watcher, error) {
	watcher := watcherLib.New()

	watcher.SetMaxEvents(1)

	watcher.FilterOps(watcherLib.Write)
	watcher.IgnoreHiddenFiles(true)

	err := watcher.AddRecursive(".")
	if err != nil {
		return nil, err
	}

	return watcher, nil
}
