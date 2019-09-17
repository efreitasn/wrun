package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"time"
)

func main() {
	// Flags
	silent := flag.Bool("s", false, "Whether the logs should be printed to stdout.")

	flag.Parse()

	cmd := exec.Command("./tt")
	cmdCompleted := make(chan error)

	if !*silent {
		cmd.Stdout = NewCmdLogger(logCmdOut)
		cmd.Stderr = NewCmdLogger(logCmdErr)
	}

	go func() {
		err := cmd.Run()
		cmdCompleted <- err
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals)

	watcher, err := createWatcher()

	if err != nil {
		logErr.Fatalln("Error while creating the watcher.")
	}

	var timer *time.Timer

	for {
		select {
		case s := <-signals:
			cmd.Process.Signal(s)
		case evt := <-watcher.Events:
			logEvt.Println(evt)

			if timer != nil {
				timer.Stop()
			}

			timer = time.AfterFunc(time.Second, func() {
				logEvt.Println("restart")
			})
		case <-watcher.Errors:
			logErr.Fatalln("Error while watching files.")
		case err := <-cmdCompleted:
			if err != nil {
				if !*silent {
					logErr.Println(fmt.Errorf("Error while running CMD: %w", err))
					logEvt.Printf("CMD exited with %v\n", cmd.ProcessState.ExitCode())
				}

				return
			}

			if !*silent {
				logEvt.Printf("CMD exited with %v\n", cmd.ProcessState.ExitCode())
			}

			return
		}
	}
}
