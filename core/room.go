package core

import (
	"net"
	"time"

	"mkw-server/logging"
	"mkw-server/settings"
)

// WFCTalkerInterface allows Room/Player to interact with WFC without circular dependency
type WFCTalkerInterface interface {
	SendPacketDataToWFC(data []byte) error
}

type Room struct {
	players map[string]*Player // key is player address string

	// UDP connection for the room, all players send/receive from this (hopefully this won't a large bottleneck with 12 players)
	conn      net.PacketConn
	addr      *net.UDPAddr
	broadcast chan Packet // channel for broadcasting packets to all players
}

var RoomInstance *Room

func InitRoom(roomAddress string) error {
	roomAddr, err := net.ResolveUDPAddr("udp", roomAddress)
	if err != nil {
		logging.Log("Failed to resolve room address %s: %v", roomAddress, err)
		return err
	}

	RoomInstance = &Room{
		players: make(map[string]*Player),
		addr:    roomAddr,
		// 256 came out of nowhere, needs to be tested
		broadcast: make(chan Packet, 256),
	}

	logging.Log("Room created successfully!", roomAddress)
	return nil
}

// Starts the listener and broadcaster goroutines
func StartRoom() {
	if RoomInstance.addr == nil {
		logging.Log("Room address is nil, cannot start room")
		return
	}

	if RoomInstance.conn != nil {
		logging.Log("Room is already started")
		return
	}

	conn, err := net.ListenPacket("udp", RoomInstance.addr.String())
	if err != nil {
		logging.Log("Failed to start room at address %s: %v", RoomInstance.addr.String(), err)
		return
	}
	logging.Log("Room listening on %s", RoomInstance.addr.String())

	RoomInstance.conn = conn

	go readLoop()
	go broadcastLoop()
}

func readLoop() {
	buf := make([]byte, 512)
	for {
		n, addr, err := RoomInstance.conn.ReadFrom(buf)
		if err != nil {
			logging.Log("Error reading from connection: %v", err)
			return
		}

		pkt := Packet{
			sender:       addr,
			data:         append([]byte{}, buf[:n]...),
			receivedTime: time.Now(),
		}

		RoomInstance.broadcast <- pkt
	}
}

func broadcastLoop() {
	for pkt := range RoomInstance.broadcast {
		for _, player := range RoomInstance.players {
			if player.addr.String() == pkt.sender.String() {
				continue
			}
			switch settings.GetPacketType() {
			case settings.CombinedRace:
				select {
				case player.sendQueue <- pkt:
				default:
					logging.Log("Send queue full for player %s, dropping packet", player.addr.String())
				}
			case settings.Race:
				_, err := RoomInstance.conn.WriteTo(pkt.data, player.addr)
				if err != nil {
					logging.Log("Error writing to player %s: %v", player.addr.String(), err)
				}
			}
		}
	}
}

func AddPlayerToRoom(playerAddr string) bool {
	if _, exists := RoomInstance.players[playerAddr]; exists {
		logging.Log("Player %s already exists in room", playerAddr)
		return false
	}

	player := NewPlayer(playerAddr, RoomInstance)
	if player == nil {
		logging.Log("Failed to create player %s", playerAddr)
		return false
	}

	RoomInstance.players[playerAddr] = player

	if settings.GetPacketType() == settings.CombinedRace {
		go player.writeLoop(RoomInstance.conn)
	}

	logging.Log("Player %s added to room", playerAddr)
	return true
}

func RemovePlayerFromRoom(playerAddr string) bool {
	p, exists := RoomInstance.players[playerAddr]
	if !exists {
		logging.Log("Player %s does not exist in room", playerAddr)
		return false
	}

	if settings.GetPacketType() == settings.CombinedRace {
		close(p.sendQueue)
	}

	delete(RoomInstance.players, playerAddr)
	logging.Log("Player %s removed from room", playerAddr)
	return true
}

func GetRoomAddr() string {
	return RoomInstance.addr.String()
}

func GetCurrentPlayerCount() int {
	return len(RoomInstance.players)
}

func CloseRoom() {
	RoomInstance.conn.Close()
}
