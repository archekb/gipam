package leaser

import (
	"errors"
	"math/rand"
	"strings"
	"sync"
	"time"

	iplib "github.com/dspinhirne/netaddr-go"
)

// NewSubnet - create new allocated block with random name (16 symbols)
func NewSubnet(pool iplib.IPNet) (*Subnet, error) {
	if pool == nil {
		return nil, errors.New("Allocated address pool is empty")
	}

	return &Subnet{ID: makeRandomString(16), Pool: pool.String(), V: uint8(pool.Version()), Idx: 1}, nil
}

// Subnet - allocated address block
type Subnet struct {
	sync.RWMutex `json:"-"`

	ID   string `json:"id"`
	V    uint8  `json:"v"`
	Pool string `json:"pool"`
	Idx  uint   `json:"idx"`

	Allocated []string `json:"allocated"`
	Free      []string `json:"free"`
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
