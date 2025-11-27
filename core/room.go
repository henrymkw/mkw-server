package core

import (
	"net"

	"mkw-server/logging"
)

type Room struct {
	players map[string]*Player // key is player address string

	// UDP connection for the room, all players send/receive from this (hopefully this won't a large bottleneck with 12 players)
	conn      net.PacketConn
	addr      *net.UDPAddr
	broadcast chan Packet // channel for broadcasting packets to all players
}

func NewRoom(roomAddress string) *Room {
	roomAddr, err := net.ResolveUDPAddr("udp", roomAddress)
	if err != nil {
		logging.Log("Failed to resolve room address %s: %v", roomAddress, err)
		return nil
	}

	r := &Room{
		players: make(map[string]*Player),
		addr:    roomAddr,
		// 256 came out of nowhere, needs to be tested
		broadcast: make(chan Packet, 256),
	}

	logging.Log("Room created successfully!", roomAddress)
	return r
}

// Starts the listener and broadcaster goroutines
func (r *Room) Start() {
	if r.addr == nil {
		logging.Log("Room address is nil, cannot start room")
		return
	}

	if r.conn != nil {
		logging.Log("Room is already started")
		return
	}

	conn, err := net.ListenPacket("udp", r.addr.String())
	if err != nil {
		logging.Log("Failed to start room at address %s: %v", r.addr.String(), err)
		return
	}
	logging.Log("Room listening on %s", r.addr.String())

	r.conn = conn

	go r.readLoop()
	go r.broadcastLoop()
}

func (r *Room) readLoop() {
	buf := make([]byte, 512)
	for {
		n, addr, err := r.conn.ReadFrom(buf)
		if err != nil {
			logging.Log("Error reading from connection: %v", err)
			return
		}

		pkt := Packet{
			sender: addr,
			data:   append([]byte{}, buf[:n]...),
		}

		// push into a broadcast channel
		r.broadcast <- pkt
	}
}

func (r *Room) broadcastLoop() {
	for pkt := range r.broadcast {
		for _, player := range r.players {
			if player.addr.String() == pkt.sender.String() {
				continue
			}
			select {
			case player.sendQueue <- pkt:
			default:
				// Packet gets dropped
			}
		}
	}
}

func (r *Room) AddPlayerToRoom(playerAddr string) bool {
	if _, exists := r.players[playerAddr]; exists {
		logging.Log("Player %s already exists in room", playerAddr)
		return false
	}

	player := NewPlayer(playerAddr, r)
	if player == nil {
		logging.Log("Failed to create player %s", playerAddr)
		return false
	}

	r.players[playerAddr] = player

	go player.writeLoop(r.conn)

	logging.Log("Player %s added to room", playerAddr)
	return true
}

func (r *Room) RemovePlayerFromRoom(playerAddr string) bool {
	p, exists := r.players[playerAddr]
	if !exists {
		logging.Log("Player %s does not exist in room", playerAddr)
		return false
	}
	// end the broadcast/goroutine for this player
	close(p.sendQueue)

	delete(r.players, playerAddr)
	logging.Log("Player %s removed from room", playerAddr)
	return true
}

func (r *Room) GetAddr() string {
	return r.addr.String()
}

func (r *Room) GetCurrentPlayerCount() int {
	return len(r.players)
}

func (r *Room) Close() {
	r.conn.Close()
}
