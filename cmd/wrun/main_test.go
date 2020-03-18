package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/efreitasn/wrun/internal/logs"
)

func TestStartCmd(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		t.Skip()
		done := make(chan string)

		go func() {
			startCmd([]string{"wrun", "init"})
			defer os.Remove("wrun.yaml")

			rp, wp := io.Pipe()
			logs.Evt.SetOutput(wp)
			r := bufio.NewReader(rp)

			go func() {
				startCmd([]string{"wrun", "start"})
			}()

			str, err := r.ReadString('\n')
			if err != nil {
				done <- fmt.Sprintf("unexpected err: %v", err)
			}

			expectedStr := logs.Evt.Prefix() + "starting cmds[0]\n"
			if str != expectedStr {
				done <- fmt.Sprintf("got %v, want %v", str, expectedStr)
			}

			if _, err := os.Create("aa.txt"); err != nil {
				done <- fmt.Sprintf("unexpected err: %v", err)
			}
			defer os.Remove("aa.txt")

			str, err = r.ReadString('\n')
			if err != nil {
				done <- fmt.Sprintf("unexpected err: %v", err)
			}

			expectedStr = logs.Evt.Prefix() + "CREATE aa.txt\n"
			if str != expectedStr {
				done <- fmt.Sprintf("got %v, want %v", str, expectedStr)
			}

			done <- ""
		}()

		select {
		case str := <-done:
			if str != "" {
				t.Error(str)
			}
		case <-time.After(2 * time.Second):
			t.Error("timeout reached")
		}
	})

	t.Run("default ingored paths", func(t *testing.T) {
		startCmd([]string{"wrun", "init"})
		defer os.Remove("wrun.yaml")

		rp, wp := io.Pipe()
		logs.Evt.SetOutput(wp)
		r := bufio.NewReader(rp)

		go func() {
			startCmd([]string{"wrun", "start"})
		}()

		strCh := make(chan struct {
			str string
			err error
		})
		go func() {
			str, err := r.ReadString('\n')
			strCh <- struct {
				str string
				err error
			}{str, err}
		}()

		select {
		case str := <-strCh:
			if str.err != nil {
				t.Fatalf("unexpected err: %v", str.err)
			}

			expectedStr := logs.Evt.Prefix() + "starting cmds[0]\n"
			if str.str != expectedStr {
				t.Fatalf("got %v, want %v", str.str, expectedStr)
			}
		case <-time.After(2 * time.Second):
			t.Fatalf("timeout reached")
		}

		if err := os.Mkdir(".git", os.ModeDir|os.ModePerm); err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		defer os.Remove(".git")

		if err := os.Mkdir(".tmp", os.ModeDir|os.ModePerm); err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		defer os.Remove(".tmp")

		if _, err := os.Create("wrun.yml"); err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		defer os.Remove("wrun.yml")

		f, err := os.OpenFile("wrun.yaml", os.O_WRONLY|os.O_APPEND, os.ModePerm)
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}

		if _, err := f.WriteString("del"); err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		f.Close()

		btCh := make(chan byte)
		go func() {
			bt, _ := r.ReadByte()
			btCh <- bt
		}()

		select {
		case bt := <-btCh:
			t.Errorf("unexpected byte: %v", bt)
		case <-time.After(2 * time.Second):
		}
	})

	t.Run("no config file", func(t *testing.T) {
		rp, wp := io.Pipe()
		logs.Err.SetOutput(wp)
		r := bufio.NewReader(rp)

		go func() {
			startCmd([]string{"wrun", "start"})
		}()

		strCh := make(chan struct {
			str string
			err error
		})
		go func() {
			str, err := r.ReadString('\n')
			strCh <- struct {
				str string
				err error
			}{str, err}
		}()

		select {
		case str := <-strCh:
			if str.err != nil {
				t.Fatalf("unexpected err: %v", str.err)
			}

			expectedStr := logs.Err.Prefix() + "config file: not found\n"
			if str.str != expectedStr {
				t.Fatalf("got %v, want %v", str.str, expectedStr)
			}
		case <-time.After(2 * time.Second):
			t.Fatalf("timeout reached")
		}
	})
}
