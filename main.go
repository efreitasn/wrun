package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	// Config
	config, err := getConfig()
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

		// A mutex is used because terminateCMD() sends a signal to an exec.Cmd and
		// exec.Cmd.Start() writes to exec.Cmd. That is also the reason why exec.Cmd.Start() and
		// exec.Cmd.Wait() were used instead of just exec.Cmd.Run().
		var cmdMx sync.Mutex

		// PRECMD
		var preCmd *exec.Cmd
		preCmdDone := make(chan struct{}, 1)

		if len(config.PreCmd) > 0 {
			preCmd = exec.CommandContext(ctx, config.PreCmd[0], config.PreCmd[1:]...)
			preCmd.Stdout = preCmdStdout
			preCmd.Stderr = preCmdStderr
		}

		// CMD
		cmd := exec.CommandContext(ctx, config.Cmd[0], config.Cmd[1:]...)
		cmdDone := make(chan struct{}, 1)

		cmd.Stdout = cmdStdout
		cmd.Stderr = cmdStderr

		go func() {
			if preCmd != nil {
				cmdMx.Lock()
				err := preCmd.Start()
				cmdMx.Unlock()

				if err != nil {
					logErr.Println(fmt.Errorf("Error while starting PRECMD: %w", err))

					if !config.SkipPreCmdErr {
						logEvt.Println("CMD will be skipped due to PRECMD error")

						return
					}
				}

				err = preCmd.Wait()

				preCmdDone <- struct{}{}

				if err != nil {
					logErr.Println(fmt.Errorf("Error while running PRECMD: %w", err))

					if !config.SkipPreCmdErr {
						logEvt.Println("CMD will be skipped due to PRECMD error")

						return
					}
				}
			}

			cmdMx.Lock()
			err := cmd.Start()
			cmdMx.Unlock()

			if err != nil {
				logErr.Println(fmt.Errorf("Error while starting CMD: %w", err))

				return
			}

			err = cmd.Wait()

			if err != nil {
				logErr.Println(fmt.Errorf("Error while running CMD: %w", err))
			}

			cmdDone <- struct{}{}
		}()

		select {
		case <-deadlySignals:
			terminateCMD(preCmd, &cmdMx, preCmdDone, cancel, config.DelayToKill)
			terminateCMD(cmd, &cmdMx, cmdDone, cancel, config.DelayToKill)
			return
		case e := <-watcher.Event:
			logEvt.Println(e)

			terminateCMD(preCmd, &cmdMx, preCmdDone, cancel, config.DelayToKill)
			terminateCMD(cmd, &cmdMx, cmdDone, cancel, config.DelayToKill)
		}
	}
}
