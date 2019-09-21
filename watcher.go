package main

import (
	watcherLib "github.com/radovskyb/watcher"
)

func createWatcher() (*watcherLib.Watcher, error) {
	watcher := watcherLib.New()

	watcher.SetMaxEvents(1)

	watcher.FilterOps(watcherLib.Create, watcherLib.Remove, watcherLib.Rename)
	watcher.AddRecursive(".")

	watcher.Ignore("./.git")
	watcher.IgnoreHiddenFiles(true)

	return watcher, nil
}
