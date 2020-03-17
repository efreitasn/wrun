package watcher

import (
	"path"
	"path/filepath"
	"strings"
)

// watchedDir represents a directory being watched.
// If it's the root, parent=nil.
type watchedDir struct {
	wd     int
	name   string
	parent *watchedDir
	// children maps name to watchedDir.
	children map[string]*watchedDir
}

type watchedDirsTreeCache struct {
	pathByWd map[int]string
	wdByPath map[string]int
}

func newWatchedDirsTreeCache() *watchedDirsTreeCache {
	return &watchedDirsTreeCache{
		pathByWd: map[int]string{},
		wdByPath: map[string]int{},
	}
}

func (wdtc *watchedDirsTreeCache) add(wd int, path string) {
	wdtc.pathByWd[wd] = path
	wdtc.wdByPath[path] = wd
}

func (wdtc *watchedDirsTreeCache) path(wd int) (string, bool) {
	path, ok := wdtc.pathByWd[wd]

	return path, ok
}

func (wdtc *watchedDirsTreeCache) wd(path string) (int, bool) {
	wd, ok := wdtc.wdByPath[path]

	return wd, ok
}

func (wdtc *watchedDirsTreeCache) rmByPath(path string) {
	wd, ok := wdtc.wd(path)
	if !ok {
		return
	}

	delete(wdtc.pathByWd, wd)
	delete(wdtc.wdByPath, path)
}

func (wdtc *watchedDirsTreeCache) rmByWd(wd int) {
	path, ok := wdtc.path(wd)
	if !ok {
		return
	}

	delete(wdtc.pathByWd, wd)
	delete(wdtc.wdByPath, path)
}

// watchedDirsTree represents a tree of watched directories starting at the working directory.
// Its methods, if incorrect data is passed to them, panic
// instead of returning errors. This happens because these
// methods aren't exposed and incorrect data must not be
// passed. Thus, returning erros would just add unnecessary
// handling. So they could either do nothing or panic. By
// panicking, invalid use of these methods can be caught
// when testing.
type watchedDirsTree struct {
	root  *watchedDir
	items map[int]*watchedDir
	cache *watchedDirsTreeCache
}

func newWatchedDirsTree() *watchedDirsTree {
	return &watchedDirsTree{
		items: map[int]*watchedDir{},
		cache: newWatchedDirsTreeCache(),
	}
}

func (wdt *watchedDirsTree) setRoot(path string, wd int) {
	if wdt.root != nil {
		panic("there's already a root")
	}

	d := &watchedDir{
		wd:       wd,
		name:     cleanPath(path),
		children: map[string]*watchedDir{},
	}

	wdt.root = d
	wdt.items[d.wd] = d
}

func (wdt *watchedDirsTree) add(wd int, name string, parentWd int) {
	parent := wdt.items[parentWd]
	if parent == nil {
		panic("parent not found")
	}

	d := &watchedDir{
		wd:       wd,
		name:     name,
		parent:   parent,
		children: map[string]*watchedDir{},
	}

	wdt.items[d.wd] = d
	d.parent.children[d.name] = d
}

func (wdt *watchedDirsTree) has(wd int) bool {
	_, ok := wdt.items[wd]

	return ok
}

func (wdt *watchedDirsTree) get(wd int) *watchedDir {
	return wdt.items[wd]
}

func (wdt *watchedDirsTree) rm(wd int) {
	item := wdt.items[wd]

	if item == nil {
		return
	}

	if item.parent == nil {
		panic("cannot remove root")
	}

	delete(item.parent.children, item.name)

	for _, child := range item.children {
		wdt.rm(child.wd)
	}

	wdt.invalidate(wd)
	delete(wdt.items, item.wd)
}

// if newParentWd < 0, the dir's parent isn't updated.
// if name == "", the dir's name isn't updated.
func (wdt *watchedDirsTree) mv(wd, newParentWd int, name string) {
	item := wdt.get(wd)
	if item == nil {
		panic("item not found")
	}

	if item.parent == nil {
		panic("cannot move root")
	}

	if newParentWd == -1 {
		newParentWd = item.parent.wd
	}

	newParent := wdt.get(newParentWd)
	if newParent == nil {
		panic("newParent not found")
	}

	if name != "" && name != item.name {
		delete(item.parent.children, item.name)
		item.name = name

		item.parent.children[name] = item
	}

	if newParentWd != item.parent.wd {
		delete(item.parent.children, item.name)
		newParent.children[item.name] = item
		item.parent = newParent
	}

	wdt.invalidate(wd)
}

func (wdt *watchedDirsTree) path(wd int) string {
	if _, ok := wdt.cache.path(wd); !ok {
		item := wdt.get(wd)
		if item == nil {
			panic("item not found while generating path")
		}

		// if this is true, it's the root
		if item.parent == nil {
			return item.name
		}

		wdt.cache.add(wd, path.Join(wdt.path(item.parent.wd), item.name))
	}

	path, _ := wdt.cache.path(wd)

	return path
}

func (wdt *watchedDirsTree) invalidate(wd int) {
	item := wdt.get(wd)
	if item == nil {
		panic("item not found")
	}

	for _, child := range item.children {
		wdt.invalidate(child.wd)
	}

	wdt.cache.rmByWd(wd)
}

func (wdt *watchedDirsTree) find(path string) *watchedDir {
	if wdt.root.name == path {
		return wdt.root
	}

	if path == "" {
		return nil
	}

	wd, ok := wdt.cache.wd(path)
	if !ok {
		pathWithoutRoot := strings.TrimPrefix(path, wdt.root.name+"/")
		pathSegments := strings.Split(pathWithoutRoot, string(filepath.Separator))

		parent := wdt.root
		for _, pathSegment := range pathSegments {
			d := parent.children[pathSegment]
			if d == nil {
				return nil
			}

			parent = d
		}

		return parent
	}

	return wdt.get(wd)
}

// cleanPath cleans the path p.
// It has the same behaviour as path.Clean(), except when p == ".",
// which results in an empty string.
func cleanPath(p string) string {
	if p == "." {
		return ""
	}

	return path.Clean(p)
}
