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
	// Config
	config, err := GetConfig()
	if err != nil {
		logErr.Println(fmt.Errorf("Error while reading config file: %w", err))

		return
	}

	// Signals
	deadlySignals := make(chan os.Signal, 1)
	signal.Notify(deadlySignals, os.Interrupt, syscall.SIGTERM)

	// Std*
	cmdStdout := NewCmdLogger(logCmdOut)
	cmdStderr := NewCmdLogger(logCmdErr)

	preCmdStdout := NewCmdLogger(logPreCmdOut)
	preCmdStderr := NewCmdLogger(logPreCmdErr)

	// Watcher
	watcher, err := CreateWatcher()
	if err != nil {
		logErr.Println(fmt.Errorf("Error while creating watcher: %w", err))

		return
	}
	// The returned error is ignored here purposely
	go watcher.Start(500 * time.Millisecond)

	for {
		ctx, cancel := context.WithCancel(context.Background())

		var preCmd *exec.Cmd
		var cmd *exec.Cmd

		if len(config.PreCmd) > 0 {
			preCmd = exec.CommandContext(ctx, config.PreCmd[0], config.PreCmd[1:]...)
			preCmd.Stdout = preCmdStdout
			preCmd.Stderr = preCmdStderr
		}

		cmd = exec.CommandContext(ctx, config.Cmd[0], config.Cmd[1:]...)
		cmd.Stdout = cmdStdout
		cmd.Stderr = cmdStderr

		go func() {
			if preCmd != nil {
				err := preCmd.Run()

				if err != nil {
					logErr.Println(fmt.Errorf("Error while running PRECMD: %w", err))
					logEvt.Println("CMD will be skipped due to PRECMD error")

					return
				}
			}

			err := cmd.Run()

			if err != nil {
				logErr.Println(fmt.Errorf("Error while running CMD: %w", err))
			}
		}()

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
