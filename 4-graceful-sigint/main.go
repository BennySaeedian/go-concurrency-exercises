//////////////////////////////////////////////////////////////////////
//
// Given is a mock process which runs indefinitely and blocks the
// program. Right now the only way to stop the program is to send a
// SIGINT (Ctrl-C). Killing a process like that is not graceful, so we
// want to try to gracefully stop the process first.
//
// Change the program to do the following:
//   1. On SIGINT try to gracefully stop the process using
//          `proc.Stop()`
//   2. If SIGINT is called again, just kill the program (last resort)
//

package main

import (
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Create a process
	proc := MockProcess{}
	// we create a buffered channel of OS signals
	signals := make(chan os.Signal, 1)
	// register the signals channel to listen to SIGINT (Ctrl-C)
	signal.Notify(signals, syscall.SIGINT)
	// run the process concurrently
	go func() {
		proc.Run()
	}()
	// wait for a signal
	<-signals
	// try to stop the program gracefully, in another concurrent goroutine
	go proc.Stop()
	// wait on the 2nd signal
	<-signals
	os.Exit(1)
}
