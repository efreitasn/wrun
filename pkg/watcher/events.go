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

// DeleteEvent represents the removal of a file or directory.
type DeleteEvent struct {
	path  string
	isDir bool
}

// IsDir returns whether the event item is a directory.
func (de DeleteEvent) IsDir() bool {
	return de.isDir
}

// Path returns the event item's path.
func (de DeleteEvent) Path() string {
	return de.path
}

// WatcherEvent returns a string representation of the event.
func (de DeleteEvent) WatcherEvent() string {
	return "remove"
}

func (de DeleteEvent) String() string {
	return de.WatcherEvent()
}

// ModifyEvent represents the removal of a file or directory.
type ModifyEvent struct {
	path  string
	isDir bool
}

// IsDir returns whether the event item is a directory.
func (me ModifyEvent) IsDir() bool {
	return me.isDir
}

// Path returns the event item's path.
func (me ModifyEvent) Path() string {
	return me.path
}

// WatcherEvent returns a string representation of the event.
func (me ModifyEvent) WatcherEvent() string {
	return "modify"
}

func (me ModifyEvent) String() string {
	return me.WatcherEvent()
}
