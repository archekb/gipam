package main

import (
	"log"

	"github.com/docker/go-plugins-helpers/ipam"
)

func init() {
}

func main() {
	cnf, err := NewConfig()
	if err != nil {
		ConfigHelp()
		log.Fatalln("Config Error:", err)
	}

	// new leaser
	leaser, err := NewLeaser(cnf.Lease.IPv6, cnf.Lease.IPv4, cnf.Lease.IPv6AB, cnf.Lease.IPv4AB)
	if err != nil {
		log.Fatalln("Create Leaser Instance Error:", err)
	}

	log.Println("Created new AddressSpace. Len main IPv6 address pool -", leaser.V6Pool.SubnetCount(leaser.V6AllocateBlock), "/ Len main IPv4 address pool -", leaser.V4Pool.SubnetCount(leaser.V4AllocateBlock))

	// Leaser Backup to file
	leaserBackup, err := NewLeaserBackup(cnf.Lease.File, leaser)
	if err != nil {
		log.Fatalln("Create GIPAM Instance Error:", err)
	}

	if !cnf.Lease.Wipe {
		err := leaserBackup.Restore()
		if err != nil {
			log.Fatalln("Restore state Error:", err)
		}
	}

	go leaserBackup.Saver()

	// new GIpam
	gipam, err := NewGIpam(leaser)
	if err != nil {
		log.Fatalln("Create GIPAM Instance Error:", err)
	}

	log.Println("Start GIPAM driver...")
	h := ipam.NewHandler(gipam)
	if cnf.Server.UnixSocket {
		log.Println(h.ServeUnix("gipam", 755))
	} else {
		log.Println(h.ServeTCP("gipam", cnf.Server.Address, "", nil))
	}
}
