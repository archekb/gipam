package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/archekb/gipam/pkg/config"
	"github.com/archekb/gipam/pkg/gipam"
	"github.com/archekb/gipam/pkg/leaser"

	"github.com/docker/go-plugins-helpers/ipam"
)

func init() {
}

func main() {
	cnf, err := config.New()
	if err != nil {
		config.Help()
		log.Fatalln("Config Error:", err)
	}

	// make new Backup to file struct
	lsrBackup, err := leaser.NewBackup(cnf.Lease.File)
	if err != nil {
		log.Fatalln("Lease Backup and Restore state from file error:", err)
	}

	// try restore previous state from file
	lsr, err := lsrBackup.Restore()
	if lsr == nil {
		log.Println("Can't Restore state from file, because", err)

		// creates new leaser if can't restore
		lsr, err = leaser.New(cnf.Lease.IPv6, cnf.Lease.IPv4, cnf.Lease.IPv6AB, cnf.Lease.IPv4AB)
		if err != nil {
			log.Fatalln("Create Leaser Instance Error:", err)
		}
	}

	// run every 30 seconds save
	ctxBackuper, cancelBackuper := context.WithCancel(context.Background())
	go lsrBackup.Saver(ctxBackuper, lsr)
	defer cancelBackuper()

	// finally save state
	defer lsrBackup.Save(lsr)

	// Notify current state
	if lsr.V6Pool != nil {
		log.Printf("IPv6 address pool: %s / Len (%d): %d", lsr.V6Pool, lsr.V6AllocateBlock, lsr.V6Pool.SubnetCount(lsr.V6AllocateBlock))
	}

	if lsr.V4Pool != nil {
		log.Printf("IPv4 address pool: %s / Len (%d): %d", lsr.V4Pool, lsr.V4AllocateBlock, lsr.V4Pool.SubnetCount(lsr.V4AllocateBlock))
	}

	// new GIpam
	g, err := gipam.New(lsr)
	if err != nil {
		log.Fatalln("Create GIPAM Instance Error:", err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		h := ipam.NewHandler(g)
		if cnf.Server.Address == "" {
			log.Println("Start UNIX socket GIPAM driver...")
			log.Println(h.ServeUnix("gipam", 755))
		} else {
			log.Println("Start TCP [" + cnf.Server.Address + "] GIPAM driver...")
			log.Println(h.ServeTCP("gipam", cnf.Server.Address, "", nil))
		}
	}()

	<-stop
	log.Println("Stop GIPAM driver")
}
