package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
)

func main() {
	if err := run(os.Args); err != nil {
		os.Exit(1)
		return
	}
}

func run(args []string) error {
	switch args[1] {
	case "run":
		parent()
	case "child":
		child()
	default:
		panic("What should i do ?")
	}
	return nil
}

func parent() {
	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags:   syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET,
		Unshareflags: syscall.CLONE_NEWNS, // Host musn't share it's namespace with container
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Println("ERROR (parent): ", err)
		os.Exit(1)
	}
}

func child() {
	fmt.Printf("Running %v as PID %d\n\n", os.Args[2:], os.Getpid())

	cntroot := "/tmp/go-fs/rootfs"
	hostroot := filepath.Join(cntroot, "/.hostroot")

	// Add CGroup
	addCGroup()

	// Create a proc directory and mount it
	mountDir(cntroot, "proc")
	must(syscall.Mount(cntroot, cntroot, "", syscall.MS_BIND|syscall.MS_REC, ""))
	must(os.MkdirAll(hostroot, 0700))
	must(syscall.PivotRoot(cntroot, hostroot))

	// Change to container's "/" point
	must(os.Chdir("/"))

	// unmount the host's "/" point
	hostroot = "/.hostroot"
	must(syscall.Unmount(hostroot, syscall.MNT_DETACH))
	// remove hostroot
	must(os.RemoveAll(hostroot))

	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	syscall.Sethostname([]byte("go-container"))

	if err := cmd.Run(); err != nil {
		fmt.Println("ERROR (child): ", err)
		os.Exit(1)
	}
}

func mountDir(cntroot string, dirname string) {
	source := dirname
	target := filepath.Join(cntroot, "/"+dirname)
	flag := 0
	fstype := dirname
	data := ""

	must(os.MkdirAll(target, 0755))
	must(syscall.Mount(source, target, fstype, uintptr(flag), data))
}

func addCGroup() {
	cgs := "/sys/fs/cgroup"
	pids := filepath.Join(cgs, "pids")
	err := os.MkdirAll(filepath.Join(pids, "go-cntr"), 0755)

	if err != nil != os.IsExist(err) {
		panic(err)
	}

	// Set max processes running in the container to be 20
	must(os.WriteFile(filepath.Join(pids, "go-cntr/pids.max"), []byte("20"), 0700))

	// Remove the new cgroup after container exits.
	must(os.WriteFile(filepath.Join(pids, "go-cntr/notify_on_release"), []byte("1"), 0700))

	must(os.WriteFile(filepath.Join(pids, "go-cntr/cgroup.procs"), []byte(strconv.Itoa(os.Getpid())), 0700))
}

func mountHost() {
	cntroot := "/tmp/go-fs/rootfs"
	hostroot := filepath.Join(cntroot, "/.hostroot")
	must(syscall.Mount(hostroot, hostroot, "", syscall.MS_BIND|syscall.MS_REC, ""))
	must(os.Chdir("~"))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
