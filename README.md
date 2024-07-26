# 117 line container

I've implemented this Go container as shown on Liz Rice's video on containerization [Container Camp](https://www.youtube.com/watch?v=Utf-A4rODH8) to deeply understand how containers really work under the hood, namespaces like the important namespace which is PID. It's 67 lines longer than the one in the video demo plus I haven't added the root filesystem management (for downloading and caching the root filesystem images that are being pivotrooted into). In this scenario, I've also used Pivotroot to swap the host and container's root, used Busybox distro as the root filesystem, though you can any distro of your preference, like CentOS, Ubuntu etc.

### Caution ⚠️⚠️
 A word of advice: It's not production-ready, and I strongly advise you not to deploy it to production. If you do deploy it, then it's on you.

### Environment 
I've tested it on my environment;

* Windows subsystem for Linux (WSL2)
* Debian 12 (Bookworm) Linux distro
* Go version 1.22.5

### How to setup 🛠️

1. Clone the Github repository to your directory and `cd` into it.
2. Run `mkdir -p /tmp/go-fs/rootfs` on your host machine and extract busybox.tar file `tar -C /tmp/go-fs/rootfs -xf assets/busybox.tar` .
3. Run `sudo ./bin/go-container run /bin/sh` in terminal to execute. `/bin/sh` is the default shell for Busybox distro. I've added a bash shell to the container so you can also run command `sudo ./bin/go-container run /bin/bash`.


After finishing the above, it shows the newly set container and it's processes.

1. Run `ps` to check the currently running processes. It should display something a bit similar to this;
![ps](/assets/ps.png)
2.  Run `ls proc/` to show its folder. It should display something similar to this;
![ls proc/](/assets/proc-folder.png)
3. Run `ls proc/mounts/` to show a new `/proc` for the container. It should display something a bit similar to this;
![/proc/mounts](/assets/proc-mounts.png)

## How to setup container to connect to host's eth0.

1. Create a network namespace
   
   `sudo ip netns add go-container-ns`

2. Run the container

   ``sudo ./bin/go-container run /bin/sh``

3. Get PID of running container from host
   `sudo lsns -t net`

4. Create a symbolic link between the network namespace and container PID

   `sudo ln -s /proc/271/ns/net /var/run/netns/go-cntr`
   
   It is a hacky solution because my host refused to create a symbolic link with the current container PID on an existing net namespace. So I created one and ran the command to create a symbolic link in the `/var/run/netns` folder.

5. Create a veth pair for host and container
   
   `sudo ip link add ve1 type veth peer name ve2`

6. Move one end of veth pair to the container namespace
   
   `sudo ip link set ve1 netns go-cntr`

7. Configure ip of container
   
   ``ip link set ve1 up``

   ``ip link set lo up``
   
   ``ip addr add 172.18.0.10/16 dev ve1``

8. Configure ip of host 
   
   `sudo ip link set ve2 up`

   `sudo ip addr add 172.18.0.20/16 dev ve2` 

9. Setup default routing in container

   `ip route add default via 192.168.1.1 `

10. Enable ip forward in host

    `sudo echo 1 > /proc/sys/net/ipv4/ip_forward`
    
     or 
    
    `sudo sysctl -w net.ipv4.ip_forward=1`

11. Set up routing in container

    `ip route add default via 172.18.0.20`

12. Setup NAT on host side to allow container to access outside world

      ``sudo iptables -t nat -A POSTROUTING -s 172.18.0.1/16 -o eth0 -j MASQUERADE``
13. Add default route to host to reach container

    `sudo ip route add 172.18.0.1/16 dev ve1`
    	
      or 

    `sudo ip route add 172.18.0.1/16 dev ve2`

14. Find the ip of host's eth0 interface

    `sudo ip addr show eth0 | grep inet`

15. Ping the host's eth0 interface from the container

    `ping -c 2 <your_host_eth0_ip>`

16. Ping every network from the container

    `ping -c 2 8.8.8.8`

     or

    `ping -c 2 google.com`


## Configuring 2 or more containers to talk to each other

#### Before all of this open 3 tabs for container 1, container 2 and host.

1. Create a network namespace
   
   `sudo ip netns add container-ns`

2. Run the container 
   
   `sudo ./bin/go-container run /bin/sh`


3. Get PID of running container's from host

   `sudo lsns -t net`

   Depending on your machine. In my case container 1 has a PID 253 and container 2 has a PID 399

   ![Containers processes](/assets/lsns.png) 

4. Create a symbolic link between the network namespace and container 1 PID <253>
   
   `sudo ln -s /proc/253/ns/net /var/run/netns/go-cntr`

5. Create a symbolic link between the network namespace and container 2 PID <399>
   
   `sudo ln -s /proc/399/ns/net /var/run/netns/go-cntr1`

   On checking all the created symbolic link container namespaces
   
     `sudo ip netns ls`

   ![Container namespaces](/assets/containers-list.png)


6. Create a veth pair for host and container 1
   
   `sudo ip link add ve1 type veth peer name ce1`

7. Create a veth pair for host and container 2
   
   `sudo ip link add ve2 type veth peer name ce2`

8. Move one end of veth pair to the container 1 namespace
   
   `sudo ip link set ce1 netns go-cntr`

9. Move one end of veth pair to the container 2 namespace
   
   `sudo ip link set ce2 netns go-cntr1`

10. Configure ip inside container 1
    
    `ip link set ce1 up`

    `ip link set lo up`

    `ip addr add 172.18.0.10/16 dev ce1`

11. Configure ip inside container 2
    
    `ip link set ce2 up`
    
    `ip link set lo up`
    
    `ip addr add 172.18.0.20/16 dev ce2`

12. Set veth devices on the host side 
    
    `sudo ip link set ve1 up && sudo ip link set ve2 up`

13. Run this command to verify that there's no new routes on the host

    `sudo ip route list`

#### Now we create a bridge to enable the containers to talk to each other like friendly neighbors.

1. Run this to create a bridge br0

   `sudo ip link add br0 type bridge`

2. Run this to set bridge up
   
   `sudo ip link set br0 up`

3. Add veth devices to the bridge
   
   `sudo ip link set ve1 master br0 && sudo ip link set ve2 master br0`

4. Try pinging container 2 from container 1
   
   `ping -c 2 172.18.0.20`

5. Do the same by pinging container 1 from container 2
   
   `ping -c 2 172.18.0.10`

6. Enable ip forward in host

   `sudo echo 1 > /proc/sys/net/ipv4/ip_forward`
    
      or
   
   `sudo sysctl -w net.ipv4.ip_forward=1`

7. Run this command in host to assign ip to bridge to establish 
   connectivity between host and container namespace.
   
   `sudo ip addr add 172.18.0.1/16 dev br0`

8. Add a default routing in container 1 and 2
   
   `ip route add default via 172.18.0.20`

9. Add a default routing in container 2

   `ip route add default via 172.18.0.20`

10. Setup NAT on the host side to allow container's to access the outside world 
    by connecting the host's to eth0.
    
    `sudo iptables -t nat -A POSTROUTING -s 172.18.0.1/16 -o eth0 -j MASQUERADE`

11. Find the ip of host's eth0 interface
    
    `sudo ip addr show eth0 | grep inet`

12. Now try pinging other networks from the container's
    
    container 1

    `ping -c 2 8.8.8.8`


    container 2

    `ping -c 2 8.8.8.8`


