package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	iplib "github.com/dspinhirne/netaddr-go"
)

// ==============
// Leaser
// ==============

// NewLeaser - create new instance of Leaser
func NewLeaser(v6, v4 string, abv6, abv4 uint) (*Leaser, error) {
	var lsr Leaser = Leaser{}

	errV6 := lsr.setV6(v6, abv6)
	if errV6 != nil {
		log.Println(errV6)
	}

	errV4 := lsr.setV4(v4, abv4)
	if errV4 != nil {
		log.Println(errV4)
	}

	if errV6 != nil && errV4 != nil {
		return nil, errors.New("Can't create new Leaser, IPv4 and IPv6 pools are empty")
	}

	return &lsr, nil
}

// Leaser - cut main IP pool to allocated blocks
type Leaser struct {
	sync.RWMutex

	V6Pool          *iplib.IPv6Net
	V6AllocateBlock uint
	V6Idx           uint

	V4Pool          *iplib.IPv4Net
	V4AllocateBlock uint
	V4Idx           uint

	Allocated []*Subnet
	Free      []*Subnet
}

// MarshalJSON -
func (lsr *Leaser) MarshalJSON() ([]byte, error) {
	lsr.Lock()
	defer lsr.Unlock()

	c := struct {
		V6Pool          string
		V6AllocateBlock uint
		V6Idx           uint

		V4Pool          string
		V4AllocateBlock uint
		V4Idx           uint

		Allocated *[]*Subnet
		Free      *[]*Subnet
	}{V6Pool: lsr.V6Pool.String(), V6AllocateBlock: lsr.V6AllocateBlock, V6Idx: lsr.V6Idx, V4Pool: lsr.V4Pool.String(), V4AllocateBlock: lsr.V4AllocateBlock, V4Idx: lsr.V4Idx, Allocated: &lsr.Allocated, Free: &lsr.Free}

	return json.MarshalIndent(c, "", "  ")
}

// UnmarshalJSON -
func (lsr *Leaser) UnmarshalJSON(data []byte) error {
	lsr.Lock()
	defer lsr.Unlock()

	c := struct {
		V6Pool          string
		V6AllocateBlock uint
		V6Idx           uint

		V4Pool           string
		V4AllocatedBlock uint
		V4Idx            uint

		Allocate *[]*Subnet
		Free     *[]*Subnet
	}{}

	err := json.Unmarshal(data, &c)
	if err != nil {
		return err
	}

	if c.Allocate != nil {
		lsr.Allocated = *c.Allocate
	}

	if c.Free != nil {
		lsr.Free = *c.Free
	}

	errV6 := lsr.setV6(c.V6Pool, c.V6AllocateBlock)
	if errV6 == nil {
		lsr.V6Idx = c.V6Idx
	}

	errV4 := lsr.setV4(c.V4Pool, c.V4AllocatedBlock)
	if errV4 == nil {
		lsr.V4Idx = c.V4Idx
	}

	if errV6 != nil && errV4 != nil {
		return errors.New("IPv4 and IPv6 pools are empty")
	}

	return nil
}

// set main IPv6 pool and allocated block
func (lsr *Leaser) setV6(pool string, ab uint) error {
	net6, err := iplib.ParseIPv6Net(pool)
	if err != nil {
		return errors.New("Can't parce main IPv6 address pool, will ignore")
	}

	switch {
	case ab >= 128:
		lsr.clearBlocks(6)
		return errors.New("Len of allocate block is 1, allocate block can't be less than 2, IPv6 will ignore")

	case net6.SubnetCount(ab) == 0:
		lsr.clearBlocks(6)
		return errors.New("Len of main IPv6 address pool to allocate block is 0, IPv6 will ignore")

	default:
		lsr.V6Pool = net6
		lsr.V6AllocateBlock = ab
	}

	return nil
}

// set main IPv4 pool and allocated block
func (lsr *Leaser) setV4(pool string, ab uint) error {
	net4, err := iplib.ParseIPv4Net(pool)
	if err != nil {
		return errors.New("Can't parce main IPv4 address pool, will ignore")
	}

	switch {
	case ab >= 32:
		lsr.clearBlocks(4)
		return errors.New("Len of allocate block is 1, allocate block can't be less than 2, IPv4 will ignore")

	case net4.SubnetCount(ab) == 0:
		lsr.clearBlocks(4)
		return errors.New("Len of main IPv4 address pool to allocate block is 0, IPv4 will ignore")

	default:
		lsr.V4Pool = net4
		lsr.V4AllocateBlock = ab
	}

	return nil
}

// clearBlocks - delete allocate and free subnet blocks by IP Version, it can be 4 or 6
func (lsr *Leaser) clearBlocks(v uint8) {
	if v != 6 && v != 4 {
		return
	}

	lsr.Lock()
	defer lsr.Unlock()

	for k, b := range lsr.Allocated {
		if b.V == v {
			lsr.Allocated[k] = lsr.Allocated[len(lsr.Allocated)-1]
			lsr.Allocated = lsr.Allocated[:len(lsr.Allocated)-1]
		}
	}

	for k, b := range lsr.Free {
		if b.V == v {
			lsr.Free[k] = lsr.Free[len(lsr.Free)-1]
			lsr.Free = lsr.Free[:len(lsr.Free)-1]
		}
	}
}

// GetBlock - get one block by IP Version, it can be 4 or 6.
// Firstly try get block from free, and if no blocks in free, cut block from main pool.
func (lsr *Leaser) GetBlock(v uint8) (string, string, error) {
	if v != 6 && v != 4 {
		return "", "", errors.New("Wrong requested IP protocol version")
	}

	lsr.Lock()
	defer lsr.Unlock()

	b, err := lsr.getBlockFromFree(v)
	if err == nil {
		lsr.Allocated = append(lsr.Allocated, b)
		return b.ID, b.Pool, nil
	}

	b, err = lsr.getBlockFromMainPool(v)
	if err == nil {
		lsr.Allocated = append(lsr.Allocated, b)
		return b.ID, b.Pool, nil
	}

	return "", "", errors.New("Have no block")
}

// get one block from available free blocks by IP Version, it can be 4 or 6.
func (lsr *Leaser) getBlockFromFree(v uint8) (*Subnet, error) {
	for k, b := range lsr.Free {
		if b.V != v {
			continue
		}
		lsr.Free[k] = lsr.Free[len(lsr.Free)-1]
		lsr.Free = lsr.Free[:len(lsr.Free)-1]

		return b, nil
	}

	return nil, errors.New("No block in Free pool")
}

// getBlockFromMainPool - cut one block from main pool by IP Version, it can be 4 or 6.
func (lsr *Leaser) getBlockFromMainPool(v uint8) (*Subnet, error) {
	switch v {
	case 6:
		if lsr.V6Pool == nil {
			return nil, errors.New("Can't get new IPv6 address block from main pool, because IPv6 block is ignore")
		}

		nbv6 := lsr.V6Pool.NthSubnet(lsr.V6AllocateBlock, uint64(lsr.V6Idx))
		if nbv6 == nil {
			return nil, errors.New("Can't get new IPv6 address block from main pool " + lsr.V6Pool.String())
		}

		lsr.V6Idx++

		b, err := NewSubnet(nbv6)
		if err != nil {
			return nil, err
		}

		return b, nil

	case 4:
		if lsr.V6Pool == nil {
			return nil, errors.New("Can't get new IPv4 address block from main pool, because IPv4 block is ignore")
		}

		nbv4 := lsr.V4Pool.NthSubnet(lsr.V4AllocateBlock, uint32(lsr.V4Idx))
		if nbv4 == nil {
			return nil, errors.New("Can't get new IPv4 address block from main pool " + lsr.V4Pool.String())
		}

		lsr.V4Idx++

		b, err := NewSubnet(nbv4)
		if err != nil {
			return nil, err
		}

		return b, nil

	default:
		return nil, errors.New("Wrong requested IP protocol version")
	}
}

// ReturnBlock - return one allocated block to free
func (lsr *Leaser) ReturnBlock(id string) error {
	lsr.Lock()
	defer lsr.Unlock()

	for k, b := range lsr.Allocated {
		if b.ID == id {
			lsr.Allocated[k] = lsr.Allocated[len(lsr.Allocated)-1]
			lsr.Allocated = lsr.Allocated[:len(lsr.Allocated)-1]
			b.Reset()
			lsr.Free = append(lsr.Free, b)

			return nil
		}
	}

	return errors.New(id + " address block not found")
}

// GetAddress - get one address from allocate block
func (lsr *Leaser) GetAddress(id string) (string, error) {
	lsr.Lock()
	defer lsr.Unlock()

	for _, b := range lsr.Allocated {
		if b.ID == id {
			ip, err := b.GetAddress()
			if err != nil {
				log.Println(err)
				return "", errors.New(id + " can't get ip address.")
			}

			switch {
			case b.V == 6:
				return ip + "/" + strconv.Itoa(int(lsr.V6AllocateBlock)), nil

			case b.V == 4:
				return ip + "/" + strconv.Itoa(int(lsr.V4AllocateBlock)), nil
			}
		}
	}

	return "", errors.New(id + " address block not found")
}

// ReturnAddress - return address to allocate block
func (lsr *Leaser) ReturnAddress(id, address string) error {
	lsr.Lock()
	defer lsr.Unlock()

	for _, b := range lsr.Allocated {
		if b.ID == id {
			return b.ReturnAddress(address)
		}
	}

	return errors.New(id + " address block not found")
}

// ==============
// Subnet
// ==============

// NewSubnet - create new allocated block with random name (16 symbols)
func NewSubnet(pool iplib.IPNet) (*Subnet, error) {
	if pool == nil {
		return nil, errors.New("Allocated address pool is empty")
	}

	return &Subnet{ID: makeRandomString(16), Pool: pool.String(), V: uint8(pool.Version()), Idx: 1}, nil
}

// Subnet - allocated address block
type Subnet struct {
	sync.RWMutex

	ID   string
	V    uint8
	Pool string
	Idx  uint

	Allocated []string
	Free      []string
}

// GetAddress - one ip from allocated address block
func (sn *Subnet) GetAddress() (string, error) {
	sn.Lock()
	defer sn.Unlock()

	ip, err := sn.getFromFreePool()
	if err == nil {
		sn.Allocated = append(sn.Allocated, ip)
		return ip, err
	}

	ip, err = sn.getFromMainPool()
	if err == nil {
		sn.Allocated = append(sn.Allocated, ip)
		return ip, err
	}

	return "", errors.New("Can't get ip from '" + sn.ID + "' allocated pool: " + sn.Pool)
}

func (sn *Subnet) getFromFreePool() (string, error) {
	if len(sn.Free) == 0 {
		return "", errors.New("No Address in free pool")
	}

	ip := sn.Free[len(sn.Free)-1]
	sn.Free = sn.Free[:len(sn.Free)-1]
	return ip, nil
}

func (sn *Subnet) getFromMainPool() (string, error) {
	switch sn.V {
	case 6:
		spv6, _ := iplib.ParseIPv6Net(sn.Pool)
		ipo := spv6.Nth(uint64(sn.Idx))
		if ipo == nil {
			return "", errors.New("No more address in '" + sn.ID + "' IPv6 Subnet pool " + sn.Pool)
		}

		sn.Idx++
		return ipo.String(), nil

	case 4:
		spv4, _ := iplib.ParseIPv4Net(sn.Pool)
		ipo := spv4.Nth(uint32(sn.Idx))
		if ipo == nil {
			return "", errors.New("No more address in '" + sn.ID + "' IPv4 Subnet pool " + sn.Pool)
		}

		sn.Idx++
		return ipo.String(), nil

	default:
		return "", errors.New("Wrong IP protocol version")
	}
}

// ReturnAddress - move ip from Allocated to Free, now IP is free and can be given in another Address request
func (sn *Subnet) ReturnAddress(ip string) error {
	// if we get ip with mask
	ip = strings.Split(ip, "/")[0]

	sn.Lock()
	defer sn.Unlock()

	for k, v := range sn.Allocated {
		if v == ip {
			sn.Allocated[k] = sn.Allocated[len(sn.Allocated)-1]
			sn.Allocated = sn.Allocated[:len(sn.Allocated)-1]
			sn.Free = append(sn.Free, ip)
			return nil
		}
	}
	return errors.New("Returned address not found in block " + sn.ID)
}

// Reset - clear
func (sn *Subnet) Reset() {
	sn.Lock()
	defer sn.Unlock()

	sn.Allocated = sn.Allocated[:0]
	sn.Free = sn.Free[:0]
	sn.Idx = 1
}

func makeRandomString(length uint) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

// ==============
// LeaserBackup
// ==============

// NewLeaserBackup - make new leaser backup
func NewLeaserBackup(lf string, lsr *Leaser) (*LeaserBackup, error) {
	if lf == "" {
		return nil, errors.New("")
	}

	if f, err := os.Stat(lf); os.IsNotExist(err) || f.IsDir() || err != nil {
		return nil, errors.New(lf + " - lease file not exist or not have permissions")
	}

	if lsr == nil {
		return nil, errors.New("")
	}

	return &LeaserBackup{LeaseFile: lf, Leaser: lsr}, nil
}

// LeaserBackup contains methods for save and restore leaser
type LeaserBackup struct {
	LeaseFile     string
	Leaser        *Leaser
	previousState *[]byte
}

// Saver - background save every 30 seconds
func (lb *LeaserBackup) Saver() {
	for {
		time.Sleep(30 * time.Second)
		lb.Save()
	}
}

// Save - save state of leases to file
func (lb *LeaserBackup) Save() {
	ml, err := lb.Leaser.MarshalJSON()
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
func (lb *LeaserBackup) Restore() error {
	lf, err := ioutil.ReadFile(lb.LeaseFile)
	if err != nil {
		return err
	}

	var l Leaser
	err = l.UnmarshalJSON(lf)
	if err != nil {
		return err
	}

	lb.Leaser = &l
	return nil
}
