package main

import (
	watcherLib "github.com/radovskyb/watcher"
)

func createWatcher(ignoreFiles []string) (*watcherLib.Watcher, error) {
	watcher := watcherLib.New()

	watcher.SetMaxEvents(1)

	watcher.FilterOps(watcherLib.Write)
	if ignoreFiles != nil {
		watcher.Ignore(ignoreFiles...)
	}

	err := watcher.AddRecursive(".")
	if err != nil {
		return nil, err
	}

	return watcher, nil
}
