package util

import (
	"encoding/binary"
	"net"
)

func CreateUDPAddr(ipUint uint32, portUint uint16) *net.UDPAddr {
	ipBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(ipBytes, ipUint)

	return &net.UDPAddr{
		IP:   net.IP(ipBytes),
		Port: int(portUint),
	}
}
