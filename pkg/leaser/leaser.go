package leaser

import (
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"sync"

	iplib "github.com/dspinhirne/netaddr-go"
)

// New creates new instance of Leaser
func New(v6, v4 string, abv6, abv4 uint) (*Leaser, error) {
	var lsr Leaser = Leaser{}

	errV6 := lsr.setV6(v6, abv6)
	if errV6 != nil {
		log.Println(errV6)
	}

	errV4 := lsr.setV4(v4, abv4)
	if errV4 != nil {
		log.Println(errV4)
		return nil, errors.New("IPv4 pool can't be empty")
	}

	return &lsr, nil
}

// Leaser - cut main IP pool to allocated blocks
type Leaser struct {
	sync.RWMutex `json:"-"`

	V6Pool          *iplib.IPv6Net `json:"v6"`
	V6AllocateBlock uint           `json:"v6ab"`
	V6Idx           uint           `json:"v6idx"`

	V4Pool          *iplib.IPv4Net `json:"v4"`
	V4AllocateBlock uint           `json:"v4ab"`
	V4Idx           uint           `json:"v4idx"`

	Allocated []*Subnet `json:"allocated,omitempty"`
	Free      []*Subnet `json:"free,omitempty"`
}

// MarshalJSON implements JSON marshaler
func (lsr *Leaser) MarshalJSON() ([]byte, error) {
	lsr.Lock()
	defer lsr.Unlock()

	c := struct {
		V6Pool          string `json:"v6"`
		V6AllocateBlock uint   `json:"v6ab"`
		V6Idx           uint   `json:"v6idx,omitempty"`

		V4Pool          string `json:"v4"`
		V4AllocateBlock uint   `json:"v4ab"`
		V4Idx           uint   `json:"v4idx,omitempty"`

		Allocated *[]*Subnet `json:"allocated,omitempty"`
		Free      *[]*Subnet `json:"free,omitempty"`
	}{V6AllocateBlock: lsr.V6AllocateBlock, V6Idx: lsr.V6Idx, V4AllocateBlock: lsr.V4AllocateBlock, V4Idx: lsr.V4Idx, Allocated: &lsr.Allocated, Free: &lsr.Free}

	if lsr.V6Pool != nil {
		c.V6Pool = lsr.V6Pool.String()
	}

	if lsr.V4Pool != nil {
		c.V4Pool = lsr.V4Pool.String()
	}

	return json.MarshalIndent(c, "", "  ")
}

// UnmarshalJSON implements JSON unmarshaler
func (lsr *Leaser) UnmarshalJSON(data []byte) error {
	lsr.Lock()
	defer lsr.Unlock()

	c := struct {
		V6Pool          string `json:"v6"`
		V6AllocateBlock uint   `json:"v6ab"`
		V6Idx           uint   `json:"v6idx,omitempty"`

		V4Pool          string `json:"v4"`
		V4AllocateBlock uint   `json:"v4ab"`
		V4Idx           uint   `json:"v4idx,omitempty"`

		Allocated *[]*Subnet `json:"allocated,omitempty"`
		Free      *[]*Subnet `json:"free,omitempty"`
	}{}

	err := json.Unmarshal(data, &c)
	if err != nil {
		return err
	}

	if c.Allocated != nil {
		lsr.Allocated = *c.Allocated
	}

	if c.Free != nil {
		lsr.Free = *c.Free
	}

	errV6 := lsr.setV6(c.V6Pool, c.V6AllocateBlock)
	if errV6 != nil {
		log.Println(errV6)
	} else {
		lsr.V6Idx = c.V6Idx
	}

	errV4 := lsr.setV4(c.V4Pool, c.V4AllocateBlock)
	if errV4 != nil {
		return errV4
	}

	lsr.V4Idx = c.V4Idx

	return nil
}

// set main IPv6 pool and allocated block
func (lsr *Leaser) setV6(pool string, ab uint) error {
	net6, err := iplib.ParseIPv6Net(pool)
	if err != nil {
		return errors.New("Can't parce main IPv6 address pool")
	}

	switch {
	case ab >= 128:
		lsr.clearBlocks(6)
		return errors.New("Len of allocate block can't be less than 2")

	case net6.SubnetCount(ab) == 0:
		lsr.clearBlocks(6)
		return errors.New("Len of main IPv6 address pool to allocate block is 0")

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
		return errors.New("Can't parce main IPv4 address pool")
	}

	switch {
	case ab >= 32:
		lsr.clearBlocks(4)
		return errors.New("Len of allocate block can't be less than 2")

	case net4.SubnetCount(ab) == 0:
		lsr.clearBlocks(4)
		return errors.New("Len of main IPv4 address pool to allocate block is 0")

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

	return "", "", err
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
