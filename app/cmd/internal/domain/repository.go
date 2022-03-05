package domain

import (
	"net"
)

type AuthData struct {
	Login    string
	Password string
	IP       net.IP
}

type BucketData struct {
	Login string
	IP    net.IP
}

type IPListRepository interface {
	Exists(addr net.IP) (bool, error)
	Add(addr *net.IPNet) (*AddressItem, error)
	Delete(addr *net.IPNet) (*AddressItem, error)
}

type IPDuplicateError struct{}

func (e *IPDuplicateError) Error() string {
	return "IP already exists"
}

type IPNotExistsError struct{}

func (e *IPNotExistsError) Error() string {
	return "IP doesn't exists"
}
