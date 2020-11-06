# gipam

Docker IPAM driver for dynamic IPv4/IPv6 allocation from Address Pool. Address pool must be routed on the host system.


### Configuration ###
---

Plugin file will be create automaticaly, if you use sudo or user on the Docker group.
* UNIX socket it will be `/run/docker/plugins/gipam.sock` [used by default]
* TCP server it will be `/etc/docker/plugins/gipam.spec`

**Attention!** Docker cache plugin file. It's mean if you start `gipam` with UNIX socket, Docker will works only with UNIX socket while you not reboot host. Look *Stop* section for more information.


**Attention!** IPv4 address pool can't be empty it's libnetwork [https://github.com/moby/libnetwork/pull/826] restriction. [https://github.com/moby/moby/issues/32850]


Enviroment variables:

* GIPAM_ADDRESS - Address and port for TCP Docker connect. If address is empty usung UNIX socket. Default: ``

* GIPAM_FILE - file for saving and restore state of driver. Default: `leases.json`
* GIPAM_V6 - Main IPv6 Address pool. Example: `2001:db8::/56`
* GIPAM_V6AB - IPv6 allocate block cutting from Main IPv6 Address pool for one service (mask). Default: `64`
* GIPAM_V4 - Main IPv4 Address pool. Example: `192.168.0.0/16`
* GIPAM_V4AB - IPv6 allocate block cutting from Main IPv4 Address pool for one service (mask). Default: `24`


Command line arguments (rewrite Enviroment variables):

* -address - Address and port for TCP Docker connect. If address is empty usung UNIX socket. Default: ``

* -file - file for saving and restore state of driver. Default: `leases.json`
* -v6 - Main IPv6 Address pool. Example: `2001:db8::/56`
* -v6ab - IPv6 allocate block cutting from Main IPv6 Address pool for one service (mask). Default: `64`
* -v4 - Main IPv4 Address pool. Example: `192.168.0.0/16`
* -v4ab - IPv6 allocate block cutting from Main IPv4 Address pool for one service (mask). Default: `24`


Lease file config (Enviroment variables and Command line interface arguments will ignored):

` {
  "v6": "2001:db8::/56",
  "v6ab": 64,
  "v6idx": 1,
  "v4": "192.168.0.0/16",
  "v4ab": 24,
  "v4idx": 1,
  "allocated": [],
  "free": []
}`


#### Tests ####
---

	go test -cover -count=1 ./...


#### Build ####
---

	go build


### Run ###
---

	sudo ./gipam -v6 2001:db8::/54 -v4 192.168.0.0/16


### Stop ###
---

	ctrl+с

If you want change connection method [UNIX, TCP] without host reboot. You need:
1) Stop all conteiners who use `gipam` driver
2) Stop `gipam` plugin
3) Delete plugin file:
* for UNIX socket `/run/docker/plugins/gipam.sock`
* for TCP server `/etc/docker/plugins/gipam.spec`
4) Restart docker daemon by `systemctl restart docker`
5) If not works, go to step 1.


### More info ###
---

API: https://docs.docker.com/engine/extend/plugin_api/

GO plugin helpers: https://github.com/docker/go-plugins-helpers