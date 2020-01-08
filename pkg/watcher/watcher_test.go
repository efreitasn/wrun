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
			t.Fatalf("unexpected error creating %v: %v", "a/b/c/d/e", err)
		}
		defer os.RemoveAll("a")

		err = os.MkdirAll("f/g/h/i/j", os.ModeDir|os.ModePerm)
		if err != nil {
			t.Fatalf("unexpected error creating %v: %v", "f/g/h/i/j", err)
		}
		defer os.RemoveAll("f")

		w, err := New([]*regexp.Regexp{})
		expectedErr := error(nil)
		if err != expectedErr {
			t.Fatalf("got %v, want %v", err, expectedErr)
		}
		defer w.Close()

		events, errs := w.Start()

		filePath := path.Join("a/b/c/d/e", "a.txt")
		_, err = os.Create(filePath)
		if err != nil {
			t.Fatalf("unexpected error creating %v: %v", filePath, err)
		}

		expectedEvent := CreateEvent{
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

	t.Run("create directory", func(t *testing.T) {
		err := os.MkdirAll("a/b/c/d/e", os.ModeDir|os.ModePerm)
		if err != nil {
			t.Fatalf("unexpected error creating %v: %v", "a/b/c/d/e", err)
		}
		defer os.RemoveAll("a")

		err = os.MkdirAll("f/g/h/i/j", os.ModeDir|os.ModePerm)
		if err != nil {
			t.Fatalf("unexpected error creating %v: %v", "f/g/h/i/j", err)
		}
		defer os.RemoveAll("f")

		w, err := New([]*regexp.Regexp{})
		expectedErr := error(nil)
		if err != expectedErr {
			t.Fatalf("got %v, want %v", err, expectedErr)
		}
		defer w.Close()

		events, errs := w.Start()

		dirPath := path.Join("a/b/c/d/e", "z")
		err = os.Mkdir(dirPath, os.ModeDir|os.ModePerm)
		if err != nil {
			t.Fatalf("unexpected error creating %v: %v", dirPath, err)
		}

		expectedEvent := CreateEvent{
			isDir: true,
			path:  dirPath,
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

		filePath := path.Join(dirPath, "a.txt")
		_, err = os.Create(filePath)
		if err != nil {
			t.Fatalf("unexpected error creating %v: %v", filePath, err)
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
			t.Fatalf("unexpected error creating %v: %v", "a/b/c/d/e", err)
		}
		defer os.RemoveAll("a")

		err = os.MkdirAll("f/g/h/i/j", os.ModeDir|os.ModePerm)
		if err != nil {
			t.Fatalf("unexpected error creating %v: %v", "f/g/h/i/j", err)
		}
		defer os.RemoveAll("f")

		w, err := New([]*regexp.Regexp{
			regexp.MustCompile("^a/b/c/d/e/a.txt$"),
		})
		expectedErr := error(nil)
		if err != expectedErr {
			t.Fatalf("got %v, want %v", err, expectedErr)
		}
		defer w.Close()

		events, errs := w.Start()

		filePath := path.Join("a/b/c/d/e", "a.txt")
		_, err = os.Create(filePath)
		if err != nil {
			t.Fatalf("unexpected error creating %v: %v", filePath, err)
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

	t.Run("create directory (regexp)", func(t *testing.T) {
		err := os.MkdirAll("a/b/c/d/e", os.ModeDir|os.ModePerm)
		if err != nil {
			t.Fatalf("unexpected error creating %v: %v", "a/b/c/d/e", err)
		}
		defer os.RemoveAll("a")

		err = os.MkdirAll("f/g/h/i/j", os.ModeDir|os.ModePerm)
		if err != nil {
			t.Fatalf("unexpected error creating %v: %v", "f/g/h/i/j", err)
		}
		defer os.RemoveAll("f")

		w, err := New([]*regexp.Regexp{
			regexp.MustCompile("^a/b/c/d/e*"),
		})
		expectedErr := error(nil)
		if err != expectedErr {
			t.Fatalf("got %v, want %v", err, expectedErr)
		}
		defer w.Close()

		events, errs := w.Start()

		dirPath := path.Join("a/b/c/d/e", "z")
		err = os.Mkdir(dirPath, os.ModeDir|os.ModePerm)
		if err != nil {
			t.Fatalf("unexpected error creating %v: %v", dirPath, err)
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

		filePath := path.Join(dirPath, "a.txt")
		_, err = os.Create(filePath)
		if err != nil {
			t.Fatalf("unexpected error creating %v: %v", filePath, err)
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

func TestWatcher_deleteEvent(t *testing.T) {
	t.Run("delete file", func(t *testing.T) {
		err := os.MkdirAll("a/b/c/d/e", os.ModeDir|os.ModePerm)
		if err != nil {
			t.Fatalf("unexpected error creating %v: %v", "a/b/c/d/e", err)
		}
		defer os.RemoveAll("a")

		err = os.MkdirAll("f/g/h/i/j", os.ModeDir|os.ModePerm)
		if err != nil {
			t.Fatalf("unexpected error creating %v: %v", "f/g/h/i/j", err)
		}
		defer os.RemoveAll("f")

		w, err := New([]*regexp.Regexp{})
		expectedErr := error(nil)
		if err != expectedErr {
			t.Fatalf("got %v, want %v", err, expectedErr)
		}
		defer w.Close()

		filePath := path.Join("a/b/c/d/e", "a.txt")
		_, err = os.Create(filePath)
		if err != nil {
			t.Fatalf("unexpected error creating %v: %v", filePath, err)
		}

		events, errs := w.Start()

		// ignore create event
		select {
		case <-events:
		case <-time.After(eventTimeout):
			t.Fatal("timeout reached waiting for event")
		}

		err = os.Remove(filePath)
		if err != nil {
			t.Fatalf("error while removing %v: %v", filePath, err)
		}

		expectedEvent := DeleteEvent{
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

	t.Run("delete directory", func(t *testing.T) {
		err := os.MkdirAll("a/b/c/d/e", os.ModeDir|os.ModePerm)
		if err != nil {
			t.Fatalf("unexpected error creating %v: %v", "a/b/c/d/e", err)
		}
		defer os.RemoveAll("a")

		err = os.MkdirAll("f/g/h/i/j", os.ModeDir|os.ModePerm)
		if err != nil {
			t.Fatalf("unexpected error creating %v: %v", "f/g/h/i/j", err)
		}
		defer os.RemoveAll("f")

		w, err := New([]*regexp.Regexp{})
		expectedErr := error(nil)
		if err != expectedErr {
			t.Fatalf("got %v, want %v", err, expectedErr)
		}
		defer w.Close()

		events, errs := w.Start()

		dirPath := path.Join("a/b/c/d/e", "z")
		err = os.Mkdir(dirPath, os.ModeDir|os.ModePerm)
		if err != nil {
			t.Fatalf("unexpected error creating %v: %v", dirPath, err)
		}

		// ignore create event
		select {
		case <-events:
		case <-time.After(eventTimeout):
			t.Fatal("timeout reached waiting for event")
		}

		err = os.RemoveAll(dirPath)
		if err != nil {
			t.Fatalf("error while removing %v: %v", dirPath, err)
		}

		expectedEvent := DeleteEvent{
			isDir: true,
			path:  dirPath,
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

	t.Run("delete file (regexp)", func(t *testing.T) {
		err := os.MkdirAll("a/b/c/d/e", os.ModeDir|os.ModePerm)
		if err != nil {
			t.Fatalf("unexpected error creating %v: %v", "a/b/c/d/e", err)
		}
		defer os.RemoveAll("a")

		err = os.MkdirAll("f/g/h/i/j", os.ModeDir|os.ModePerm)
		if err != nil {
			t.Fatalf("unexpected error creating %v: %v", "f/g/h/i/j", err)
		}
		defer os.RemoveAll("f")

		filePath := path.Join("f/g/h/i/j", "foo")
		_, err = os.Create(filePath)
		if err != nil {
			t.Fatalf("unexpected error creating %v: %v", filePath, err)
		}

		w, err := New([]*regexp.Regexp{
			regexp.MustCompile("^" + filePath + "$"),
		})
		expectedErr := error(nil)
		if err != expectedErr {
			t.Fatalf("got %v, want %v", err, expectedErr)
		}
		defer w.Close()

		events, errs := w.Start()

		err = os.Remove(filePath)
		if err != nil {
			t.Fatalf("error while removing %v: %v", filePath, err)
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

	t.Run("delete directory (regexp)", func(t *testing.T) {
		err := os.MkdirAll("a/b/c/d/e", os.ModeDir|os.ModePerm)
		if err != nil {
			t.Fatalf("unexpected error creating %v: %v", "a/b/c/d/e", err)
		}
		defer os.RemoveAll("a")

		err = os.MkdirAll("f/g/h/i/j", os.ModeDir|os.ModePerm)
		if err != nil {
			t.Fatalf("unexpected error creating %v: %v", "f/g/h/i/j", err)
		}
		defer os.RemoveAll("f")

		dirPath := path.Join("f/g/h/i/j", "foo")
		_, err = os.Create(dirPath)
		if err != nil {
			t.Fatalf("unexpected error creating %v: %v", dirPath, err)
		}

		w, err := New([]*regexp.Regexp{
			regexp.MustCompile("^" + dirPath + "$"),
		})
		expectedErr := error(nil)
		if err != expectedErr {
			t.Fatalf("got %v, want %v", err, expectedErr)
		}
		defer w.Close()

		events, errs := w.Start()

		err = os.RemoveAll(dirPath)
		if err != nil {
			t.Fatalf("error while removing %v: %v", dirPath, err)
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
