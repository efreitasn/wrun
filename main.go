package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Watcher
	watcher, err := createWatcher()
	if err != nil {
		logErr.Println(fmt.Errorf("Error while creating watcher: %w", err))
	}

	// Signals
	deadlySignals := make(chan os.Signal, 1)
	signal.Notify(deadlySignals, os.Interrupt, syscall.SIGTERM)

	cmdStdout := NewCmdLogger(logCmdOut)
	cmdStderr := NewCmdLogger(logCmdErr)

	go watcher.Start(time.Millisecond)

	for {
		ctx, cancel := context.WithCancel(context.Background())
		cmd := exec.CommandContext(ctx, "bash", "script.sh")
		cmd.SysProcAttr = &syscall.SysProcAttr{
			// https://medium.com/@felixge/killing-a-child-process-and-all-of-its-children-in-go-54079af94773
			Setpgid: true,
		}

		cmd.Stdout = cmdStdout
		cmd.Stderr = cmdStderr

		cmd.Start()

		// ss:
		select {
		case <-deadlySignals:
			cancel()
			return
		case e := <-watcher.Event:
			logEvt.Println(e)

			cancel()
		}
	}
}
