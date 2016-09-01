package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"
)

var (
	provisionScriptPath = "/var/pcfdev/run"
	timeoutDuration     = "1h"
)

type timeoutError struct{}

func (t *timeoutError) Error() string {
	return "timeout error"
}

func main() {
	cmd := exec.Command(provisionScriptPath, os.Args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	exitCodeChan := make(chan int, 1)
	errChan := make(chan error, 1)

	go func() {
		duration, err := time.ParseDuration(timeoutDuration)
		if err != nil {
			panic(err)
		}

		<-time.After(duration)
		exitCodeChan <- 1
		errChan <- &timeoutError{}
	}()

	go func() {
		err := cmd.Run()
		exitStatus := 0

		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
					exitStatus = status.ExitStatus()
				} else {
					exitStatus = 1
				}
			} else {
				exitStatus = 1
			}
		}

		exitCodeChan <- exitStatus
		errChan <- err
	}()

	if err := <-errChan; err != nil {
		fmt.Printf("Error: %s.", err)
		os.Exit(<-exitCodeChan)
	}

}
