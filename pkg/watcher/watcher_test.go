package watcher

import (
	"os"
	"path"
	"regexp"
	"testing"
	"time"
)

// eventTimeout represents the amount of time to wait for an event.
var eventTimeout = time.Millisecond * 100

func TestWatcher_createEvent(t *testing.T) {
	t.Run("create file", func(t *testing.T) {
		err := os.MkdirAll("a/b/c/d/e", os.ModeDir|os.ModePerm)
		if err != nil {
			t.Fatalf("error while creating %v: %v", "a/b/c/d/e", err)
		}
		defer os.RemoveAll("a")

		err = os.MkdirAll("f/g/h/i/j", os.ModeDir|os.ModePerm)
		if err != nil {
			t.Fatalf("error while creating %v: %v", "f/g/h/i/j", err)
		}
		defer os.RemoveAll("f")

		w, err := New([]*regexp.Regexp{})
		expectedErr := error(nil)
		if err != expectedErr {
			t.Fatalf("got %v, want %v", err, expectedErr)
		}

		filePath := path.Join("a/b/c/d/e", "a.txt")
		_, err = os.Create(filePath)
		if err != nil {
			t.Fatalf("error while creating %v: %v", filePath, err)
		}

		expectedEvent := CreateEvent{
			isDir: false,
			path:  filePath,
		}
		events, errs := w.Start()

		select {
		case e := <-events:
			if e != expectedEvent {
				t.Fatalf("got %v, want %v", e, expectedEvent)
			}
		case err := <-errs:
			t.Fatalf("unexpected err: %v", err)
		case <-w.Done:
			t.Fatal("channel closed")
		case <-time.After(eventTimeout):
			t.Fatal("timeout reached waiting for event")
		}
	})

	t.Run("create directory", func(t *testing.T) {
		err := os.MkdirAll("a/b/c/d/e", os.ModeDir|os.ModePerm)
		if err != nil {
			t.Fatalf("error while creating %v: %v", "a/b/c/d/e", err)
		}
		defer os.RemoveAll("a")

		err = os.MkdirAll("f/g/h/i/j", os.ModeDir|os.ModePerm)
		if err != nil {
			t.Fatalf("error while creating %v: %v", "f/g/h/i/j", err)
		}
		defer os.RemoveAll("f")

		w, err := New([]*regexp.Regexp{})
		expectedErr := error(nil)
		if err != expectedErr {
			t.Fatalf("got %v, want %v", err, expectedErr)
		}

		dirPath := path.Join("a/b/c/d/e", "z")
		err = os.Mkdir(dirPath, os.ModeDir|os.ModePerm)
		if err != nil {
			t.Fatalf("error while creating %v: %v", dirPath, err)
		}

		expectedEvent := CreateEvent{
			isDir: true,
			path:  dirPath,
		}
		events, errs := w.Start()

		select {
		case e := <-events:
			if e != expectedEvent {
				t.Fatalf("got %v, want %v", e, expectedEvent)
			}
		case err := <-errs:
			t.Fatalf("unexpected err: %v", err)
		case <-w.Done:
			t.Fatal("channel closed")
		case <-time.After(eventTimeout):
			t.Fatal("timeout reached waiting for event")
		}

		filePath := path.Join(dirPath, "a.txt")
		_, err = os.Create(filePath)
		if err != nil {
			t.Fatalf("error while creating %v: %v", filePath, err)
		}

		expectedEvent = CreateEvent{
			isDir: false,
			path:  filePath,
		}

		select {
		case e := <-events:
			if e != expectedEvent {
				t.Fatalf("got %v, want %v", e, expectedEvent)
			}
		case err := <-errs:
			t.Fatalf("unexpected err: %v", err)
		case <-w.Done:
			t.Fatal("channel closed")
		case <-time.After(eventTimeout):
			t.Fatal("timeout reached waiting for event")
		}
	})

	t.Run("create file (regexp)", func(t *testing.T) {
		err := os.MkdirAll("a/b/c/d/e", os.ModeDir|os.ModePerm)
		if err != nil {
			t.Fatalf("error while creating %v: %v", "a/b/c/d/e", err)
		}
		defer os.RemoveAll("a")

		err = os.MkdirAll("f/g/h/i/j", os.ModeDir|os.ModePerm)
		if err != nil {
			t.Fatalf("error while creating %v: %v", "f/g/h/i/j", err)
		}
		defer os.RemoveAll("f")

		w, err := New([]*regexp.Regexp{
			regexp.MustCompile("^a/b/c/d/e/a.txt$"),
		})
		expectedErr := error(nil)
		if err != expectedErr {
			t.Fatalf("got %v, want %v", err, expectedErr)
		}

		filePath := path.Join("a/b/c/d/e", "a.txt")
		_, err = os.Create(filePath)
		if err != nil {
			t.Fatalf("error while creating %v: %v", filePath, err)
		}

		events, errs := w.Start()

		select {
		case e := <-events:
			t.Fatalf("unexpected event %v", e)
		case err := <-errs:
			t.Fatalf("unexpected err: %v", err)
		case <-w.Done:
			t.Fatal("channel closed")
		case <-time.After(eventTimeout):
		}
	})

	t.Run("create directory (regexp)", func(t *testing.T) {
		err := os.MkdirAll("a/b/c/d/e", os.ModeDir|os.ModePerm)
		if err != nil {
			t.Fatalf("error while creating %v: %v", "a/b/c/d/e", err)
		}
		defer os.RemoveAll("a")

		err = os.MkdirAll("f/g/h/i/j", os.ModeDir|os.ModePerm)
		if err != nil {
			t.Fatalf("error while creating %v: %v", "f/g/h/i/j", err)
		}
		defer os.RemoveAll("f")

		w, err := New([]*regexp.Regexp{
			regexp.MustCompile("^a/b/c/d/e*"),
		})
		expectedErr := error(nil)
		if err != expectedErr {
			t.Fatalf("got %v, want %v", err, expectedErr)
		}

		dirPath := path.Join("a/b/c/d/e", "z")
		err = os.Mkdir(dirPath, os.ModeDir|os.ModePerm)
		if err != nil {
			t.Fatalf("error while creating %v: %v", dirPath, err)
		}

		events, errs := w.Start()

		select {
		case e := <-events:
			t.Fatalf("unexpected event %v", e)
		case err := <-errs:
			t.Fatalf("unexpected err: %v", err)
		case <-w.Done:
			t.Fatal("channel closed")
		case <-time.After(eventTimeout):
		}

		filePath := path.Join(dirPath, "a.txt")
		_, err = os.Create(filePath)
		if err != nil {
			t.Fatalf("error while creating %v: %v", filePath, err)
		}

		select {
		case e := <-events:
			t.Fatalf("unexpected event %v", e)
		case err := <-errs:
			t.Fatalf("unexpected err: %v", err)
		case <-w.Done:
			t.Fatal("channel closed")
		case <-time.After(eventTimeout):
		}
	})
}
