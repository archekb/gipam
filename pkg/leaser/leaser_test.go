package leaser

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	cases := map[string]struct {
		Success bool
		NetV6   string
		NetV6AB uint
		NetV4   string
		NetV4AB uint
		ErrStr  string
	}{
		"Wrong v6 and v4 address pool and mask":        {NetV6: "bla", NetV6AB: 0, NetV4: "bla", NetV4AB: 0, Success: false, ErrStr: "IPv4 address pool can't be empty"},
		"Empty v6 and v4 address pool and mask":        {NetV6: "", NetV6AB: 0, NetV4: "", NetV4AB: 0, Success: false, ErrStr: "IPv4 address pool can't be empty"},
		"Wrong v6 and v4 allocate block (AB)":          {NetV6: "fe80::/48", NetV6AB: 0, NetV4: "192.168.0.1/16", NetV4AB: 0, Success: false, ErrStr: "Can't create new Leaser, IPv4 and IPv6 pools are empty"},
		"Empty v6 and v4 address pool, but AB is ok":   {NetV6: "", NetV6AB: 64, NetV4: "", NetV4AB: 24, Success: false, ErrStr: "IPv4 address pool can't be empty"},
		"Right v6 pool and wrong AB=128, empty v4":     {NetV6: "fe80::/48", NetV6AB: 128, NetV4: "", NetV4AB: 0, Success: false, ErrStr: "Can't create new Leaser, IPv4 and IPv6 pools are empty"},
		"Right v6 pool and wrong AB=129, empty v4":     {NetV6: "fe80::/48", NetV6AB: 129, NetV4: "", NetV4AB: 0, Success: false, ErrStr: "Can't create new Leaser, IPv4 and IPv6 pools are empty"},
		"Right v6 pool (64) and wrong AB=56, empty v4": {NetV6: "fe80::/64", NetV6AB: 56, NetV4: "", NetV4AB: 0, Success: false, ErrStr: "Can't create new Leaser, IPv4 and IPv6 pools are empty"},
		"Right v4 pool and wrong AB=32, empty v6":      {NetV6: "", NetV6AB: 0, NetV4: "192.168.0.1/16", NetV4AB: 32, Success: false, ErrStr: "Can't create new Leaser, IPv4 and IPv6 pools are empty"},
		"Right v4 pool and wrong AB=33, empty v6":      {NetV6: "", NetV6AB: 0, NetV4: "192.168.0.1/16", NetV4AB: 33, Success: false, ErrStr: "Can't create new Leaser, IPv4 and IPv6 pools are empty"},
		"Right v4 pool (16) and wrong AB=8, empty v6":  {NetV6: "", NetV6AB: 0, NetV4: "192.168.0.1/16", NetV4AB: 8, Success: false, ErrStr: "Can't create new Leaser, IPv4 and IPv6 pools are empty"},

		"Right v6 and empty v4": {NetV6: "fe80::/48", NetV6AB: 64, NetV4: "", NetV4AB: 0, Success: true},
		"Empty v6 and right v4": {NetV6: "", NetV6AB: 0, NetV4: "192.168.0.1/16", NetV4AB: 24, Success: true},
		"Right v6 and v4":       {NetV6: "fe80::/48", NetV6AB: 64, NetV4: "192.168.0.1/16", NetV4AB: 24, Success: true},
	}

	for k, v := range cases {
		t.Run(k, func(t *testing.T) {
			t.Parallel()
			lsr, err := New(v.NetV6, v.NetV4, v.NetV6AB, v.NetV4AB)
			if !v.Success {
				require.Nil(t, lsr)
				require.EqualError(t, err, v.ErrStr)
			} else {
				require.NotNil(t, lsr)
				require.NoError(t, err)
			}

		})
	}
}

func TestGetBlock(t *testing.T) {
	lsr, err := New("fe80::/48", "192.168.0.1/16", 64, 24)
	if err != nil {
		t.Error("Expected success create new Leaser")
	}

	name, ipnet, err := lsr.GetBlock(4)
	if len(name) == 0 || ipnet != "192.168.0.0/24" || err != nil {
		t.Error("Expected success for get IPv4 block")
	}

	name, ipnet, err = lsr.GetBlock(6)
	if len(name) == 0 || ipnet != "fe80::/64" || err != nil {
		t.Error("Expected success for get IPv6 block")
	}

	name, ipnet, err = lsr.GetBlock(0)
	if len(name) != 0 || len(ipnet) != 0 || err == nil {
		t.Error("Expected fail for IPv0 (error version of ip protocol), but success")
	}
}

func TestReturnBlock(t *testing.T) {
	lsr, err := New("fe80::/48", "192.168.0.1/16", 64, 24)
	if lsr == nil || err != nil {
		t.Error("Expected success create new Leaser")
	}

	name, ipnet, err := lsr.GetBlock(4)
	if len(name) == 0 || ipnet != "192.168.0.0/24" || err != nil {
		t.Error("Expected success for get IPv4 block")
	}

	err = lsr.ReturnBlock(name)
	if err != nil {
		t.Error("Expected success for return IPv4 block")
	}

	name, ipnet, err = lsr.GetBlock(6)
	if len(name) == 0 || ipnet != "fe80::/64" || err != nil {
		t.Error("Expected success for get IPv6 block")
	}

	err = lsr.ReturnBlock(name)
	if err != nil {
		t.Error("Expected success for return IPv6 block")
	}

	err = lsr.ReturnBlock("aaa")
	if err == nil {
		t.Error("Expected fail for return unknown 'aaa' block")
	}
}

func TestGetAddress(t *testing.T) {
	lsr, err := New("fe80::/48", "192.168.0.1/16", 64, 24)
	if lsr == nil || err != nil {
		t.Error("Expected success create new Leaser")
	}

	name, ipnet, err := lsr.GetBlock(4)
	if len(name) == 0 || ipnet != "192.168.0.0/24" || err != nil {
		t.Error("Expected success for get IPv4 block")
	}

	addr, err := lsr.GetAddress(name)
	if addr != "192.168.0.1/24" || err != nil {
		t.Error("Expected success for get IPv4 address from block")
	}

	name, ipnet, err = lsr.GetBlock(6)
	if len(name) == 0 || ipnet != "fe80::/64" || err != nil {
		t.Error("Expected success for get IPv6 block")
	}

	addr, err = lsr.GetAddress(name)
	if addr != "fe80::1/64" || err != nil {
		t.Error("Expected success for get IPv6 address from block")
	}

	addr, err = lsr.GetAddress("aaa")
	if len(addr) != 0 || err == nil {
		t.Error("Expected fail for get IP address from unknown 'aaa' block")
	}
}

func TestReturnAddress(t *testing.T) {
	lsr, err := New("fe80::/48", "192.168.0.1/16", 64, 24)
	if lsr == nil || err != nil {
		t.Error("Expected success create new Leaser")
	}

	name, ipnet, err := lsr.GetBlock(4)
	if len(name) == 0 || ipnet != "192.168.0.0/24" || err != nil {
		t.Error("Expected success for get IPv4 block")
	}

	addr, err := lsr.GetAddress(name)
	if addr != "192.168.0.1/24" || err != nil {
		t.Error("Expected success for get IPv4 address from block")
	}

	err = lsr.ReturnAddress(name, addr)
	if err != nil {
		t.Error("Expected success for return IPv4 address to block")
	}

	err = lsr.ReturnAddress(name, "123.12.23.13")
	if err == nil {
		t.Error("Expected fail for return IPv4 address to block")
	}

	name, ipnet, err = lsr.GetBlock(6)
	if len(name) == 0 || ipnet != "fe80::/64" || err != nil {
		t.Error("Expected success for get IPv6 block")
	}

	addr, err = lsr.GetAddress(name)
	if addr != "fe80::1/64" || err != nil {
		t.Error("Expected success for get IPv6 address from block")
	}

	err = lsr.ReturnAddress(name, addr)
	if err != nil {
		t.Error("Expected success for return IPv6 address to block")
	}

	err = lsr.ReturnAddress(name, "fe80::100/64")
	if err == nil {
		t.Error("Expected fail for return IPv6 address to block")
	}

	err = lsr.ReturnAddress("aaa", "fe80::900/64")
	if err == nil {
		t.Error("Expected fail for return IP address to 'aaa' unknown block")
	}
}
