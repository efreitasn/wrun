package watcher

import "testing"

import "path"

func TestWatchedDirsTreeSetRoot(t *testing.T) {
	wdt := newWatchedDirsTree()

	rootWd := 0
	wdt.setRoot(rootWd)

	if wdt.root == nil {
		t.Fatalf("got %v, want %v", nil, "non-nil value")
	}

	if wdt.root.wd != rootWd {
		t.Fatalf("got %v, want %v", wdt.root.wd, rootWd)
	}
}

func TestWatchedDirsTreeAddHasGet(t *testing.T) {
	wdt := newWatchedDirsTree()
	wdt.setRoot(0)

	dirWd := 1
	dirName := "some"

	wdt.add(dirWd, dirName, wdt.root.wd)

	has := wdt.has(dirWd)
	if !has {
		t.Errorf("got %v, want %v", false, true)
	}

	d := wdt.get(dirWd)

	if d == nil {
		t.Fatalf("got %v, want %v", nil, "non-nil value")
	}

	if d.wd != dirWd {
		t.Errorf("got %v, want %v", d.wd, dirWd)
	}

	if d.name != dirName {
		t.Errorf("got %v, want %v", d.name, dirName)
	}

	if d.parent != wdt.root {
		t.Errorf("got %v, want %v", d.parent, wdt.root)
	}
}

func TestWatchedDirsTreeRmHasGet(t *testing.T) {
	wdt := newWatchedDirsTree()
	wdt.setRoot(0)

	dirWd := 1
	dirName := "some"

	wdt.add(dirWd, dirName, wdt.root.wd)
	wdt.rm(dirWd)

	has := wdt.has(dirWd)
	if has {
		t.Errorf("got %v, want %v", true, false)
	}

	d := wdt.get(dirWd)

	if d != nil {
		t.Fatalf("got %v, want %v", d, nil)
	}
}

func TestWatchedDirsTreeRmHasGet_child(t *testing.T) {
	wdt := newWatchedDirsTree()
	wdt.setRoot(0)

	dirWd := 1
	dirName := "some"
	parentWd := 10

	wdt.add(parentWd, "foo", wdt.root.wd)
	wdt.add(dirWd, dirName, parentWd)
	wdt.rm(parentWd)

	has := wdt.has(dirWd)
	if has {
		t.Errorf("got %v, want %v", true, false)
	}

	d := wdt.get(dirWd)

	if d != nil {
		t.Fatalf("got %v, want %v", d, nil)
	}
}

func TestWatchedDirsTreeMvInvalidate(t *testing.T) {
	wdt := newWatchedDirsTree()
	wdt.setRoot(0)

	dir1Wd := 1
	dir1Name := "some"
	dir1ParentWd := wdt.root.wd
	dir2Wd := 2
	dir2Name := "foo"
	dir2ParentWd := dir1Wd
	dir3Wd := 3
	dir3Name := "bar"
	dir3ParentWd := dir1Wd
	dir4Wd := 4
	dir4Name := "fourth"
	dir4ParentWd := dir3Wd

	wdt.add(dir1Wd, dir1Name, dir1ParentWd)
	wdt.add(dir2Wd, dir2Name, dir2ParentWd)
	wdt.add(dir3Wd, dir3Name, dir3ParentWd)
	wdt.add(dir4Wd, dir4Name, dir4ParentWd)

	// add entry to cache
	wdt.path(dir4Wd)

	wdt.mv(dir4Wd, dir2Wd)

	/*
		before moving:
		- root
			- dir1
				- dir2
				- dir3
					- dir4

		after moving:
		- root
			- dir1
				- dir2
					- dir4
				- dir3
	*/

	expectedDir4Path := path.Join(
		dir1Name,
		dir2Name,
		dir4Name,
	)
	dir4Path := wdt.path(dir4Wd)

	if dir4Path != expectedDir4Path {
		t.Errorf("got %v, want %v", dir4Path, expectedDir4Path)
	}
}

func TestWatchedDirsTreePath(t *testing.T) {
	wdt := newWatchedDirsTree()
	wdt.setRoot(0)

	dir1Wd := 1
	dir1Name := "some"
	dir1ParentWd := wdt.root.wd
	dir2Wd := 2
	dir2Name := "foo"
	dir2ParentWd := dir1Wd
	dir3Wd := 3
	dir3Name := "bar"
	dir3ParentWd := dir2Wd

	wdt.add(dir1Wd, dir1Name, dir1ParentWd)
	wdt.add(dir2Wd, dir2Name, dir2ParentWd)
	wdt.add(dir3Wd, dir3Name, dir3ParentWd)

	expectedDir3Path := path.Join(
		dir1Name,
		dir2Name,
		dir3Name,
	)
	dir3Path := wdt.path(dir3Wd)

	if dir3Path != expectedDir3Path {
		t.Errorf("got %v, want %v", dir3Path, expectedDir3Path)
	}
}

func TestWatchedDirsTreeFind(t *testing.T) {
	wdt := newWatchedDirsTree()
	wdt.setRoot(0)

	dir1Wd := 1
	dir1Name := "some"
	dir1ParentWd := wdt.root.wd
	dir2Wd := 2
	dir2Name := "foo"
	dir2ParentWd := dir1Wd
	dir3Wd := 3
	dir3Name := "bar"
	dir3ParentWd := dir2Wd

	wdt.add(dir1Wd, dir1Name, dir1ParentWd)
	wdt.add(dir2Wd, dir2Name, dir2ParentWd)
	wdt.add(dir3Wd, dir3Name, dir3ParentWd)

	// the path was created this way instead of using
	// wdt.path() so that a cache entry wouldn't be
	// created.
	dir3Path := path.Join(
		dir1Name,
		dir2Name,
		dir3Name,
	)
	dir3 := wdt.get(dir3Wd)
	findRes := wdt.find(dir3Path)

	if findRes != dir3 {
		t.Errorf("got %v, want %v", findRes, dir3)
	}
}

func TestWatchedDirsTreeFind_notFound(t *testing.T) {
	wdt := newWatchedDirsTree()
	wdt.setRoot(0)

	dir1Wd := 1
	dir1Name := "some"
	dir1ParentWd := wdt.root.wd
	dir2Wd := 2
	dir2Name := "foo"
	dir2ParentWd := dir1Wd
	dir3Wd := 3
	dir3Name := "bar"
	dir3ParentWd := dir2Wd

	wdt.add(dir1Wd, dir1Name, dir1ParentWd)
	wdt.add(dir2Wd, dir2Name, dir2ParentWd)
	wdt.add(dir3Wd, dir3Name, dir3ParentWd)

	// the path was created this way instead of using
	// wdt.path() so that a cache entry wouldn't be
	// created.
	dir3Path := path.Join(
		dir1Name,
		dir2Name,
		dir3Name,
	)
	findRes := wdt.find(path.Join(dir3Path, "aa"))

	if findRes != nil {
		t.Errorf("got %v, want %v", findRes, nil)
	}
}
