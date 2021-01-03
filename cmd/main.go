package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
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
		Cloneflags:   syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
		Unshareflags: syscall.CLONE_NEWNS,
	}

	must(cmd.Run())
}

func child() {
	fmt.Printf("Running %v as %d\n", os.Args[2:], os.Getpid())

	cg()

	syscall.Sethostname([]byte("container"))
	syscall.Chroot("/var/lib/lxc/buildos7/rootfs")
	syscall.Chdir("/")
	syscall.Mount("proc", "proc", "proc", 0, "")

	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Run()

	syscall.Unmount("/proc", 0)

}

func cg() {
	cgroups := "/sys/fs/cgroup"
	scope := filepath.Join(cgroups, "system.slice/andrey-container.scope")
	err := os.Mkdir(scope, 0755)
	if err != nil && !os.IsExist(err) {
		panic(err)
	}
	must(ioutil.WriteFile(filepath.Join(scope, "pids.max"), []byte("20"), 0700))
	// Removes the new cgroup in place after the container exists
	// must(ioutil.WriteFile(filepath.Join(scope, "cgroup.events"), []byte("populated 1"), 0700))
	must(ioutil.WriteFile(filepath.Join(scope, "cgroup.procs"), []byte(strconv.Itoa(os.Getpid())), 0700))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
