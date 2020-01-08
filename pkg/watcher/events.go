package watcher

import (
	"fmt"
)

// Event is an event emitted by a watcher.
type Event interface {
	fmt.Stringer
	WatcherEvent() string
	IsDir() bool
	Path() string
}

// CreateEvent represents the creation of a file or directory.
type CreateEvent struct {
	path  string
	isDir bool
}

// IsDir returns whether the event item is a directory.
func (ce CreateEvent) IsDir() bool {
	return ce.isDir
}

// Path returns the event item's path.
func (ce CreateEvent) Path() string {
	return ce.path
}

// WatcherEvent returns a string representation of the event.
func (ce CreateEvent) WatcherEvent() string {
	return "create"
}

func (ce CreateEvent) String() string {
	return ce.WatcherEvent()
}
