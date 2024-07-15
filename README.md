# 91 line container

I've implemented this Go container as shown on Liz Rice's video on containerization [Container Camp](https://www.youtube.com/watch?v=Utf-A4rODH8) to deeply understand how containers really work under the hood, namespaces like the important namespace which is PID. It's 41 lines longer than the one in the video demo plus I haven't added the root filesystem management (for downloading and caching the root filesystem images that are being pivotrooted into). In this scenario, I've also used Pivotroot to swap the host and container's root, used Busybox distro as the root filesystem, though you can any distro of your preference, like CentOS, Ubuntu etc.

### Caution ⚠️⚠️
 A word of advice: It's not production-ready, and I strongly advise you not to deploy it to production. If you do deploy it, then it's on you.

### Environment 
I've tested it on my environment;

* Windows subsystem of Linux (WSL2)
* Debian 12 (Bookworm) Linux distro
* Go version 1.22.5

### How to setup 🛠️

1. Clone the Github repository to your directory and `cd` into it.
2. Run `mkdir -p /tmp/go-fs/rootfs` on your host machine and extract busybox.tar file `tar -C /tmp/go-fs/rootfs -xf assets/busybox.tar` .
3. Run `sudo ./bin/go-container run /bin/sh` in terminal to execute. `/bin/sh` is the default shell for Busybox distro.


After finishing the above, it shows the newly set container and it's processes.

1. Run `ps` to check the currently running processes. It should display something a bit similar to this;
![ps](/assets/ps.png)
2.  Run `ls proc/` to show its folder. It should display something similar to this;
![ls proc/](/assets/proc-folder.png)
3. Run `ls proc/mounts/` to show a new `/proc` for the container. It should display something a bit similar to this;
![/proc/mounts](/assets/proc-mounts.png)





