package main

import(
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"path/filepath"
)

func main(){
	switch os.Args[1] {
	case "run":
		parent()
	case "child":
		child()
	default:
		panic("What should i do ?")
	}
}

func parent(){
	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Println("ERROR (parent): ", err)
		os.Exit(1)
	}
}

func child(){
	fmt.Printf("Running %v as PID %d\n\n", os.Args[2:], os.Getpid())

	cntroot := "/tmp/go-fs/rootfs"
	hostroot := filepath.Join(cntroot,"/.hostroot")

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

	if err := cmd.Run(); err != nil {
		fmt.Println("ERROR (child): ", err)
		os.Exit(1)
	}
}

func mountDir(cntroot string, dirname string){
	source := dirname
	target := filepath.Join(cntroot, "/"+dirname)
	flag := 0
	fstype := dirname
	data := ""

	must(os.MkdirAll(target, 0755))
	must(syscall.Mount(source, target, fstype, uintptr(flag), data))
}

func mountHost(){
	cntroot := "/tmp/go-fs/rootfs"
	hostroot := filepath.Join(cntroot,"/.hostroot")
	must(syscall.Mount(hostroot, hostroot, "", syscall.MS_BIND|syscall.MS_REC, ""))
	must(os.Chdir("~"))
}

func must(err error){
	if err != nil {
		panic(err)
	}
}
