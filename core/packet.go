package core

import (
	"net"
	"time"
)

type Packet struct {
	sender       net.Addr
	data         []byte
	receivedTime time.Time
}
