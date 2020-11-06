package gipam

import (
	"testing"
)

type testLeaser struct{}

func (t *testLeaser) GetBlock(v uint8) (string, string, error) {
	return "aaa", "192.168.1.0/16", nil
}

func (t *testLeaser) ReturnBlock(string) error {
	return nil
}

func (t *testLeaser) GetAddress(string) (string, error) {
	return "", nil
}

func (t *testLeaser) ReturnAddress(string, string) error {
	return nil
}

// fail make GIpam new leaser is nil
func TestNewGIpam1(t *testing.T) {
	gipam, err := New(nil)
	if gipam != nil || err == nil {
		t.Error("Expected fail for init with nil")
	}
}

// success new GIpam
func TestNewGIpam2(t *testing.T) {
	gipam, err := New(&testLeaser{})
	if gipam == nil || err != nil {
		t.Error("Expected success for init with correct leaser")
	}
}

func TestGetCapabilities(t *testing.T) {
	t.Skip()
}

func TestGetDefaultAddressSpaces(t *testing.T) {
	t.Skip()
}

func TestRequestPool(t *testing.T) {
	t.Skip()
}

func TestReleasePool(t *testing.T) {
	t.Skip()
}

func TestRequestAddress(t *testing.T) {
	t.Skip()
}

func TestReleaseAddress(t *testing.T) {
	t.Skip()
}
