package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
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

	go func() { watch() }()
	go func() {
		err := cmd.Run()
		cmdCompleted <- err
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals)

	for {
		select {
		case s := <-signals:
			cmd.Process.Signal(s)
		case err := <-cmdCompleted:
			if err != nil {
				if !*silent {
					logErr.Print(fmt.Errorf("Error while running CMD: %w", err))
					logEvt.Print(
						fmt.Sprintf("CMD exited with %v", cmd.ProcessState.ExitCode()),
					)
				}

				return
			}

			if !*silent {
				logEvt.Print(
					fmt.Sprintf("CMD exited with %v", cmd.ProcessState.ExitCode()),
				)
			}

			return
		}
	}
}
