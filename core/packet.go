package core

import (
	"net"
)

type Packet struct {
	sender net.Addr
	data   []byte
}
