package main

import (
	"errors"
	"log"

	"github.com/docker/go-plugins-helpers/ipam"
)

// NewGIpam - create new instance of GIpam
func NewGIpam(leaser LeaserInterface) (*GIpam, error) {
	if leaser == nil {
		return nil, errors.New("Leaser interface is empty")
	}

	return &GIpam{Leaser: leaser}, nil
}

// LeaserInterface - interface implements internal logic
type LeaserInterface interface {
	GetBlock(uint8) (string, string, error)
	ReturnBlock(string) error
	GetAddress(string) (string, error)
	ReturnAddress(string, string) error
}

// GIpam - implement Docker IPAM Interface
type GIpam struct {
	Leaser LeaserInterface
}

// GetCapabilities - returns driver capabilities [ whether or not this IPAM required pre-made MAC ]
func (i *GIpam) GetCapabilities() (*ipam.CapabilitiesResponse, error) {
	return &ipam.CapabilitiesResponse{RequiresMACAddress: true}, nil
}

// GetDefaultAddressSpaces - returns the default local and global address space names for this IPAM
func (i *GIpam) GetDefaultAddressSpaces() (*ipam.AddressSpacesResponse, error) {
	return &ipam.AddressSpacesResponse{}, nil
}

// RequestPool - get one allocated block of IP addresses for lease it to containers
func (i *GIpam) RequestPool(request *ipam.RequestPoolRequest) (*ipam.RequestPoolResponse, error) {
	if request.V6 {
		id, ip, err := i.Leaser.GetBlock(6)
		log.Println("RequestPool:", id, ip)

		return &ipam.RequestPoolResponse{PoolID: id, Pool: ip, Data: nil}, err
	}

	id, ip, err := i.Leaser.GetBlock(4)
	log.Println("RequestPool:", id, ip)

	return &ipam.RequestPoolResponse{PoolID: id, Pool: ip, Data: nil}, err
}

// ReleasePool - return allocated block of IP addresses
func (i *GIpam) ReleasePool(request *ipam.ReleasePoolRequest) error {
	log.Println("ReleasePool:", request.PoolID)
	err := i.Leaser.ReturnBlock(request.PoolID)

	return err
}

// RequestAddress - get one address from allocated block
func (i *GIpam) RequestAddress(request *ipam.RequestAddressRequest) (*ipam.RequestAddressResponse, error) {
	// check request for gateway address and return gateway address if true
	gw, ok := request.Options["RequestAddressType"]
	if ok && gw == "com.docker.network.gateway" {
		ip, err := i.Leaser.GetAddress(request.PoolID)
		log.Println("RequestAddress:", ip, err)

		return &ipam.RequestAddressResponse{Address: ip, Data: nil}, err
	}

	ip, err := i.Leaser.GetAddress(request.PoolID)
	log.Println("RequestAddress:", ip, err)

	return &ipam.RequestAddressResponse{Address: ip, Data: nil}, err
}

// ReleaseAddress - return address to block
func (i *GIpam) ReleaseAddress(request *ipam.ReleaseAddressRequest) error {
	log.Println("ReleaseAddress:", request.PoolID, request.Address)
	err := i.Leaser.ReturnAddress(request.PoolID, request.Address)

	return err
}
