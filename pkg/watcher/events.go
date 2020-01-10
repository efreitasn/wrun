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
	str := fmt.Sprintf("CREATE %v", ce.path)

	return str
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
	str := fmt.Sprintf("DELETE %v", de.path)

	return str
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
	str := fmt.Sprintf("MODIFY %v", me.path)

	return str
}

func (me ModifyEvent) String() string {
	return me.WatcherEvent()
}

// RenameEvent represents the moving of a file or directory.
type RenameEvent struct {
	// OldPath can be equal to "" if the old path is from an unwatched directory.
	OldPath string
	path    string
	isDir   bool
}

// IsDir returns whether the event item is a directory.
func (re RenameEvent) IsDir() bool {
	return re.isDir
}

// Path returns the event item's path.
// Path can be equal to "" if the new path is from an unwatched directory.
func (re RenameEvent) Path() string {
	return re.path
}

// WatcherEvent returns a string representation of the event.
func (re RenameEvent) WatcherEvent() string {
	var str string

	switch {
	case re.OldPath != "" && re.path != "":
		str = fmt.Sprintf("RENAME %v to %v", re.OldPath, re.path)
	case re.OldPath != "":
		str = fmt.Sprintf("RENAME %v", re.OldPath)
	case re.path != "":
		str = fmt.Sprintf("RENAME to %v", re.path)
	}

	return str
}

func (re RenameEvent) String() string {
	return re.WatcherEvent()
}
