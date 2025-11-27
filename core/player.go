package core

import (
	"net"

	"mkw-server/logging"
)

type Player struct {
	conn        net.UDPConn  // Address players send and receive from
	addr        *net.UDPAddr // The resolved UDP address of this player
	roomPointer *Room        // The room this player is in

	sendQueue chan Packet
	// We have a few different options on how to structure communication
	// - I chose to have one goroutine per room to listen and read incoming packets
	//   - maybe this could be a bottleneck
	// - One goroutine for looping over the broadcast channel, which adds the packet to each player's send queue
	// - One goroutine per player to write packets from their send queue to their address
}

func NewPlayer(addr string, room *Room) *Player {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		logging.Log("Failed to resolve player address %s: %v", addr, err)
		return nil
	}

	player := &Player{
		addr:      udpAddr,
		sendQueue: make(chan Packet, 32),
	}

	return player
}

func (p *Player) writeLoop(conn net.PacketConn) {
	for pkt := range p.sendQueue {
		conn.WriteTo(pkt.data, p.addr)
	}
}

func (p *Player) SetRoomPointer(room *Room) {
	p.roomPointer = room
}

func (p *Player) GetAddr() string {
	return p.addr.String()
}
