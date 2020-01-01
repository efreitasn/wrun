package main

import (
	"context"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Config
	config, err := getConfig()

	if err != nil {
		logErr.Printf("config file: %v\n", err)

		return
	}

	globMatches, err := getGlobMatches(config)
	if err != nil {
		logErr.Printf("config file: %v\n", err)

		return
	}

	// Signals
	deadlySignals := make(chan os.Signal, 1)
	signal.Notify(deadlySignals, os.Interrupt, syscall.SIGTERM)

	// Watcher
	watcher, err := createWatcher(globMatches)
	if err != nil {
		logErr.Printf("watcher: %v\n", err)

		return
	}

	// The returned error is ignored here purposely
	go watcher.Start(500 * time.Millisecond)

	for {
		// allCmdsForCurrentEvtCtx is used to indicate that all cmds related to the current event
		// must be terminated as soon as possible.
		allCmdsForCurrentEvtCtx, cancelAllCmdsForCurrentEvtCtx := context.WithCancel(context.Background())
		// allCmdsForCurrentEvtDone indicates that all cmds related to the current event have completed
		// or been terminated.
		allCmdsForCurrentEvtDone := make(chan struct{})

		go func() {
			for i, cmdItem := range config.cmds {
				logEvt.Printf("starting cmds[%v]\n", i)
				err := runCmd(allCmdsForCurrentEvtCtx, cmdItem)

				if err != nil {
					logErr.Printf("cmds[%v]: %v\n", i, err)

					if cmdItem.fatalIfErr {
						logEvt.Println("the remaining cmds will be skipped due to the fatalIfErr flag")

						break
					}

					continue
				}
			}

			close(allCmdsForCurrentEvtDone)
		}()

		select {
		case <-deadlySignals:
			select {
			case <-allCmdsForCurrentEvtDone:
				// not really necessary
				cancelAllCmdsForCurrentEvtCtx()
			default:
				cancelAllCmdsForCurrentEvtCtx()
				<-allCmdsForCurrentEvtDone
			}

			return
		case e := <-watcher.Event:
			logEvt.Println(e)

			cancelAllCmdsForCurrentEvtCtx()
			<-allCmdsForCurrentEvtDone
		}
	}
}

// runCmd runs the given cmd.
// ctx -> indicates that the cmd must be terminated as soon as possible.
// cmdCtx -> indicates that the cmd must be terminated immediately.
// cmdDone -> indicates that the cmd has completed or been terminated.
func runCmd(ctx context.Context, cmd cmd) error {
	cmdCtx, killCmd := context.WithCancel(context.Background())
	defer killCmd()
	cmdDone := make(chan error)

	cmdExec := exec.CommandContext(cmdCtx, cmd.terms[0], cmd.terms[1:]...)

	outPipe, err := cmdExec.StdoutPipe()
	if err != nil {
		return err
	}
	go logCmdStd(cmdCtx, logCmdOut, outPipe)

	errPipe, err := cmdExec.StderrPipe()
	if err != nil {
		return err
	}
	go logCmdStd(cmdCtx, logCmdErr, errPipe)

	err = cmdExec.Start()
	if err != nil {
		return err
	}

	go func() {
		cmdDone <- cmdExec.Wait()
		close(cmdDone)
	}()

	select {
	case <-ctx.Done():
		cmdExec.Process.Signal(os.Interrupt)

		timer := time.NewTimer(time.Duration(int(time.Millisecond) * cmd.delayToKill))

		select {
		case <-timer.C:
			killCmd()

			if err = <-cmdDone; err != nil {
				return err
			}
		case err = <-cmdDone:
			return err
		}
	case err := <-cmdDone:
		return err
	}

	return nil
}

func logCmdStd(ctx context.Context, l *log.Logger, std io.Reader) {
	bs := make([]byte, 4096)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		n, err := std.Read(bs)
		if err != nil {
			return
		}

		nBs := bs[0:n]

		if nBs[len(nBs)-1] == '\n' {
			l.Print(string(nBs))
		} else {
			l.Println(string(bs))
		}
	}
}
