package util

import (
	"encoding/binary"
	"fmt"
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

func FormatIPPort(ipAddr uint32, port uint16) string {
	ipBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(ipBytes, ipAddr)

	host := net.IP(ipBytes).String()

	return net.JoinHostPort(host, fmt.Sprintf("%d", port))
}
