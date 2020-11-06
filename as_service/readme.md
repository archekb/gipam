### Description ###
---

Create linux service by systemd with GIPAM driver for Docker.


### Setup Service ###
---

1.    Copy dir `as_service` from root of repo to opt and rename it to `gipam`
2.    Build `gipam` and copy it to `/opt/gipam/`
3.    Configure `gipam` by set Enviroment variables in file `/opt/gipam/.env`
4.    Create symlink and setup service:

	sudo su
	ln -s /opt/gipam/gipam.service /etc/systemd/system/gipam.service
	systemctl enable gipam
	systemctl start gipam


### Delete Service ###
---

1.    Stop `gipam`, disable service, delete symlink: 

	sudo su
	systemctl stop gipam
	systemctl disable gipam
	rm /etc/systemd/system/gipam.service

2.    Delete folder `/opt/gipam/`


### More info ###
---

API: https://docs.docker.com/engine/extend/plugin_api/#systemd-socket-activation