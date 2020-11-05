package leaser

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"time"
)

// NewBackup makes new leaser backup.
// It will save leser structure to file and restore it.
func NewBackup(file string) (*Backup, error) {
	if file == "" {
		return nil, errors.New("file name is empty")
	}

	return &Backup{LeaseFile: file, previousState: &[]byte{}}, nil
}

// Backup contains methods for save and restore leaser
type Backup struct {
	LeaseFile     string
	previousState *[]byte
}

// Saver - background save every 30 seconds
func (lb *Backup) Saver(ctx context.Context, lsr *Leaser) {
	for {
		select {
		case <-ctx.Done():
			log.Println("Saver canseled")
			return

		case <-time.NewTicker(30 * time.Second).C:
			lb.Save(lsr)
		}
	}
}

// Save - save state of leases to file
func (lb *Backup) Save(lsr *Leaser) {
	ml, err := lsr.MarshalJSON()
	if err != nil {
		log.Println("Create json with leases error:", err)
		return
	}

	if bytes.Equal(*lb.previousState, ml) {
		return
	}

	file, err := os.Create(lb.LeaseFile)
	if err != nil {
		log.Println("Open leases file error:", err)
		return
	}
	defer file.Close()

	_, err = file.Write(ml)
	if err != nil {
		log.Println("Write leases file error:", err)
	} else {
		lb.previousState = &ml
	}
}

// Restore - restore previous state from file
func (lb *Backup) Restore() (*Leaser, error) {
	lf, err := ioutil.ReadFile(lb.LeaseFile)
	if err != nil {
		return nil, err
	}

	var lsr Leaser
	err = lsr.UnmarshalJSON(lf)
	if err != nil {
		return nil, err
	}

	return &lsr, nil
}
