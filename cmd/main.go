package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
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
	fmt.Printf("Running %v as %d, uid: %d\n", os.Args[2:], os.Getpid(), os.Getuid())

	cmd := exec.Command("/proc/self/exe", append([]string{"child", strconv.Itoa(os.Getuid())}, os.Args[2:]...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWUSER,
		Credential: &syscall.Credential{Uid: 0, Gid: 0},
		UidMappings: []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: os.Getuid(), Size: 1},
		},
		GidMappings: []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: os.Getgid(), Size: 1},
		},
	}

	must(cmd.Run())
}

func child() {
	fmt.Printf("Running %v as %d, uid: %s --> %d\n", os.Args[3:], os.Getpid(), os.Args[2], os.Getuid())

	cgUser(os.Args[2])

	syscall.Sethostname([]byte("container"))
	syscall.Chroot("/var/lib/lxc/buildos7/rootfs")
	syscall.Chdir("/")
	syscall.Mount("proc", "proc", "proc", 0, "")

	cmd := exec.Command(os.Args[3], os.Args[4:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Run()

	syscall.Unmount("/proc", 0)

}

func cgRoot() {
	cgroups := "/sys/fs/cgroup"
	scope := filepath.Join(cgroups, "system.slice/andrey-container.scope")
	err := os.Mkdir(scope, 0755)
	if err != nil && !os.IsExist(err) {
		panic(err)
	}
	must(ioutil.WriteFile(filepath.Join(scope, "pids.max"), []byte("20"), 0700))
	must(ioutil.WriteFile(filepath.Join(scope, "cgroup.procs"), []byte(strconv.Itoa(os.Getpid())), 0700))
}

func cgUser(uid string) {
	cgroups := "/sys/fs/cgroup"
	path := fmt.Sprintf("user.slice/user-%s.slice/user@%s.service/container.scope", uid, uid)
	scope := filepath.Join(cgroups, path)
	err := os.Mkdir(scope, 0755)
	if err != nil && !os.IsExist(err) {
		panic(err)
	}
	must(ioutil.WriteFile(filepath.Join(scope, "pids.max"), []byte("20"), 0700))
	must(ioutil.WriteFile(filepath.Join(scope, "cgroup.procs"), []byte(strconv.Itoa(os.Getpid())), 0700))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
