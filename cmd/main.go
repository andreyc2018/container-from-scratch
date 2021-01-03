package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"syscall"
)

// docker         run image <cmd> <args>
// go run main.go run       <cmd> <args>

func main() {
	if len(os.Args) < 3 {
		fmt.Printf("\"run <cmd> [args]\" is required.\n")
		os.Exit(0)
	}

	user, err := user.Current()
	must(err)
	if user.Uid != "0" {
		fmt.Printf("Must be root to run this.\n")
		os.Exit(0)
	}

	switch os.Args[1] {
	case "run":
		run()

	case "child":
		child()

	default:
		panic("unknown command")
	}
}

func run() {
	fmt.Printf("Running %v as %d\n", os.Args[2:], os.Getpid())

	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID,
	}

	must(cmd.Run())
}

func child() {
	fmt.Printf("Running %v as %d\n", os.Args[2:], os.Getpid())

	syscall.Sethostname([]byte("container"))
	syscall.Chroot("/var/lib/lxc/buildos7/rootfs")
	syscall.Chdir("/")

	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	must(cmd.Run())
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
