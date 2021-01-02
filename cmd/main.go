package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// docker         run image <cmd> <args>
// go run main.go run       <cmd> <args>

func main() {
	if len(os.Args) < 3 {
		fmt.Printf("\"run <cmd> [args]\" is required.\n")
		os.Exit(0)
	}

	switch os.Args[1] {
	case "run":
		run()

	default:
		panic("unknown command")
	}
}

func run() {
	fmt.Printf("Running %v\n", os.Args[2:])

	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS,
	}

	must(cmd.Run())

}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
