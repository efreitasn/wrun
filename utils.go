package main

import (
	"os"
	"os/exec"
	"sync"
	"time"
)

func terminateCMD(cmd *exec.Cmd, mx *sync.Mutex, done chan struct{}, killFn func()) {
	mx.Lock()
	if cmd == nil || cmd.Process == nil {
		mx.Unlock()
		return
	}

	cmd.Process.Signal(os.Interrupt)
	mx.Unlock()

	timer := time.NewTimer(time.Second * 10)

	select {
	case <-done:
		return
	case <-timer.C:
		killFn()
		return
	}
}
