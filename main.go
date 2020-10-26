package main

import (
	"log"

	"github.com/archekb/gipam/lib/config"
	"github.com/archekb/gipam/lib/gipam"
	"github.com/archekb/gipam/lib/leaser"

	"github.com/docker/go-plugins-helpers/ipam"
)

func init() {
}

func main() {
	cnf, err := config.NewConfig()
	if err != nil {
		config.ConfigHelp()
		log.Fatalln("Config Error:", err)
	}

	// new leaser
	lsr, err := leaser.NewLeaser(cnf.Lease.IPv6, cnf.Lease.IPv4, cnf.Lease.IPv6AB, cnf.Lease.IPv4AB)
	if err != nil {
		log.Fatalln("Create Leaser Instance Error:", err)
	}

	log.Println("Created new AddressSpace. Len main IPv6 address pool -", lsr.V6Pool.SubnetCount(lsr.V6AllocateBlock), "/ Len main IPv4 address pool -", lsr.V4Pool.SubnetCount(lsr.V4AllocateBlock))

	// Leaser Backup to file
	lsrBackup, err := leaser.NewLeaserBackup(cnf.Lease.File, lsr)
	if err != nil {
		log.Fatalln("Create GIPAM Instance Error:", err)
	}

	if !cnf.Lease.Wipe {
		err := lsrBackup.Restore()
		if err != nil {
			log.Fatalln("Restore state Error:", err)
		}
	}

	go lsrBackup.Saver()

	// new GIpam
	g, err := gipam.NewGIpam(lsr)
	if err != nil {
		log.Fatalln("Create GIPAM Instance Error:", err)
	}

	log.Println("Start GIPAM driver...")
	h := ipam.NewHandler(g)
	if cnf.Server.UnixSocket {
		log.Println(h.ServeUnix("gipam", 755))
	} else {
		log.Println(h.ServeTCP("gipam", cnf.Server.Address, "", nil))
	}
}
