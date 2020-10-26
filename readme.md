# gipam

Docker IPAM driver for dynamic IPv4/IPv6 allocation from Address Pool. Address pool must be routed on the host system.


### Configuration ###
---

Plugin file in `/etc/docker/plugins` will be create automaticaly if you use sudo or your user in Docker group. 

Enviroment variables (First stage):
* GIPAM_UNIX - use UNIX socket for connecting to Docker. Default: `false`
* GIPAM_ADDRESS - Address and port for Docker connect. Default: `:9090`

* GIPAM_FILE - file for saving and restore state of driver. Default: `leases.json`

* GIPAM_V6 - Main IPv6 Address pool. Example: `fe80::/56`
* GIPAM_V6AB - IPv6 allocate block cutting from Main IPv6 Address pool for one service (mask). Default: `64`
* GIPAM_V4 - Main IPv4 Address pool. Example: `192.168.0.0/16`
* GIPAM_V4AB - IPv6 allocate block cutting from Main IPv4 Address pool for one service (mask). Default: `24`

Command line arguments (Second stage rewrite First stage):
* -unix - use UNIX socket for connecting to Docker. Default: `false`
* -address - Address and port for Docker connect (need configure in Docker too). Default: `:9090`

* -file - file for saving and restore state of driver. Default: `leases.json`
* 
* -v6 - Main IPv6 Address pool. Example: `fe80::/56`
* -v6ab - IPv6 allocate block cutting from Main IPv6 Address pool for one service (mask). Default: `64`
* -v4 - Main IPv4 Address pool. Example: `192.168.0.0/16`
* -v4ab - IPv6 allocate block cutting from Main IPv4 Address pool for one service (mask). Default: `24`

Lease file config (Third stage rewrite First and Second stage):
` {
  "V6Pool": "2a02:17d0::/56",
  "V6AllocateBlock": 64,
  "V6Idx": 1,
  "V4Pool": "192.168.0.0/16",
  "V4AllocateBlock": 24,
  "V4Idx": 1,
  "Allocated": [],
  "Free": []
}`


#### Build ####
---

	go build


### Run ###
---

	sudo ./gipam -unix -v6 2a02:17d0::/56 -v4 10.2.0.0/16

Need *sudo* or your user must be in Docker group, because creates plugin file: `/etc/docker/plugins`


### Stop ###
---

	ctrl+с


### Set as Service ###
---

1.    Copy dir `as_service` from root of repo to opt and rename it to `gipam`
2.    Build `gipam` and copy it to `/opt/gipam/`
3.    Configure `gipam` by set Enviroment variables in file `/opt/gipam/.env`
4.    Create symlink and setup service:

	sudo su
	ln -s /opt/gipam/gipam.service /etc/systemd/system/gipam.service
	systemctl enable gipam
	systemctl start gipam

**Attention!** Application save data to `/opt/gipam/leases.json` and creates plugin file in `/etc/docker/plugins`, for this operations need user with right permissions (or *root*).