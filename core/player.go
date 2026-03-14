package core

import (
	"net"
	"time"

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
		addr:      	udpAddr,
		sendQueue: 	make(chan Packet, 32),
	}

	return player
}

func (p *Player) writeLoop(conn net.PacketConn) {
	// will need to configure and experiment with the time and batch sizes, etc
	ticker := time.NewTicker(500 * time.Microsecond)
	defer ticker.Stop()

	const maxUDPSize = 1465

	// Map from sender address to packet
	batch := make(map[string]Packet, 12)
	var oldestTime time.Time

	for {
		select {
		case packet, ok := <-p.sendQueue:
			if !ok {
				return
			}
			senderAddr := packet.sender.String()
			// Update with latest packet from this sender
			batch[senderAddr] = packet
			// Track oldest packet time
			if oldestTime.IsZero() || packet.receivedTime.Before(oldestTime) {
				oldestTime = packet.receivedTime
			}
		case <-ticker.C:
			if len(batch) == 0 {
				continue
			}

			packetsToSend := make([][]byte, 0, len(batch))

			// Collect packets to send (one per sender)
			for _, pkt := range batch {
				packetsToSend = append(packetsToSend, pkt.data)
			}

			// Build combined packet with header
			combinedPackets := BuildCombinedPacket(packetsToSend, maxUDPSize)

			// Send all combined packets
			for _, combined := range combinedPackets {
				conn.WriteTo(combined, p.addr)
				header := ParseCombinedPacketHeader(combined)
				logging.Log("Sent combined packet: %d packets, %d bytes to %s, delay: %v",
					header.NumPackets, len(combined), p.addr.String(), time.Since(oldestTime))
			}

			// Clear batch
			batch = make(map[string]Packet, 12)
			oldestTime = time.Time{}
		}
	}
}

func (p *Player) SetRoomPointer(room *Room) {
	p.roomPointer = room
}

func (p *Player) GetAddr() string {
	return p.addr.String()
}
