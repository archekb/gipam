package leaser

import (
	"testing"
)

func TestNewLeaser(t *testing.T) {
	variants := []struct {
		Success bool
		NetV6   string
		NetV6AB uint
		NetV4   string
		NetV4AB uint
	}{
		{NetV6: "bla", NetV6AB: 0, NetV4: "bla", NetV4AB: 0, Success: false},
		{NetV6: "", NetV6AB: 0, NetV4: "", NetV4AB: 0, Success: false},
		{NetV6: "fe80::/48", NetV6AB: 0, NetV4: "192.168.0.1/16", NetV4AB: 0, Success: false},
		{NetV6: "", NetV6AB: 64, NetV4: "", NetV4AB: 24, Success: false},
		{NetV6: "fe80::/48", NetV6AB: 128, NetV4: "", NetV4AB: 24, Success: false},
		{NetV6: "fe80::/48", NetV6AB: 129, NetV4: "", NetV4AB: 24, Success: false},
		{NetV6: "fe80::/64", NetV6AB: 56, NetV4: "", NetV4AB: 24, Success: false},
		{NetV6: "", NetV6AB: 0, NetV4: "192.168.0.1/16", NetV4AB: 32, Success: false},
		{NetV6: "", NetV6AB: 0, NetV4: "192.168.0.1/16", NetV4AB: 33, Success: false},
		{NetV6: "", NetV6AB: 0, NetV4: "192.168.0.1/16", NetV4AB: 8, Success: false},

		{NetV6: "fe80::/48", NetV6AB: 64, NetV4: "", NetV4AB: 24, Success: true},
		{NetV6: "", NetV6AB: 0, NetV4: "192.168.0.1/16", NetV4AB: 24, Success: true},
		{NetV6: "fe80::/48", NetV6AB: 64, NetV4: "192.168.0.1/16", NetV4AB: 24, Success: true},
	}

	for k, v := range variants {
		lsr, err := NewLeaser(v.NetV6, v.NetV4, v.NetV6AB, v.NetV4AB)
		if !v.Success && (lsr != nil || err == nil) {
			t.Errorf("Test %d, expected fail, but success", k)
		} else if v.Success && (lsr == nil || err != nil) {
			t.Errorf("Test %d, expected success, but fail", k)
		}
	}
}

func TestGetBlock(t *testing.T) {
	lsr, err := NewLeaser("fe80::/48", "192.168.0.1/16", 64, 24)
	if lsr == nil || err != nil {
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
	lsr, err := NewLeaser("fe80::/48", "192.168.0.1/16", 64, 24)
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
	lsr, err := NewLeaser("fe80::/48", "192.168.0.1/16", 64, 24)
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
	lsr, err := NewLeaser("fe80::/48", "192.168.0.1/16", 64, 24)
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
