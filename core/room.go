package core

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"

	"mkw-server/logging"
	"mkw-server/settings"
	"mkw-server/util"
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

	aidBitmap uint32
}

var room *Room

func InitRoom(roomAddr *net.UDPAddr) error {
	room = &Room{
		players: make(map[string]*Player),
		addr:    roomAddr,
		// 256 came out of nowhere, needs to be tested
		broadcast: make(chan Packet, 256),
	}

	logging.Log("Room created successfully!", roomAddr.String())
	return nil
}

// Starts the listener and broadcaster goroutines
func StartRoom() {
	if room.addr == nil {
		logging.Log("Room address is nil, cannot start room")
		return
	}

	if room.conn != nil {
		logging.Log("Room is already started")
		return
	}

	conn, err := net.ListenPacket("udp", room.addr.String())
	if err != nil {
		logging.Log("Failed to start room at address %s: %v", room.addr.String(), err)
		return
	}
	logging.Log("Room listening on %s", room.addr.String())

	room.conn = conn

	go readLoop()
	go broadcastLoop()
}

func readLoop() {
	buf := make([]byte, 512)
	for {
		n, addr, err := room.conn.ReadFrom(buf)
		if err != nil {
			logging.Log("Error reading from connection: %v", err)
			return
		}

		pkt := Packet{
			sender:       addr,
			data:         append([]byte{}, buf[:n]...),
			receivedTime: time.Now(),
		}

		room.broadcast <- pkt
	}
}



func broadcastLoop() {
	for pkt := range room.broadcast {
		data := pkt.data
		sendersAid := data[1]

		if sendersAid != room.players[pkt.sender.String()].aid {
			continue
		}

		aidBitmap := binary.BigEndian.Uint16(data[2:4])
		receivingAids := util.GetSendToAids(aidBitmap)

		for _, aid := range *receivingAids {
			p := getPlayer(aid) 
			if p == nil {
				logging.Log("Unable to find player for aid %d", aid)
				continue
			}

			if aid == sendersAid {
				logging.Log("Attempted to send from %d to %d", sendersAid, aid)
				continue
			}

			_, err := room.conn.WriteTo(pkt.data, p.addr)
			if err != nil {
				logging.Log("Error writitng to aid %d", p.aid)
				continue
			}
		}
	}
}

func getPlayer(aid byte) *Player {
	for addr, p := range room.players {
		if addr == "" || p == nil {
			continue
		}
		
		if p.aid == aid {
			return p
		}
	}
	return nil
}

func AddPlayerToRoom(playerAddr string, aid byte) error {
	if _, exists := room.players[playerAddr]; exists {
		return fmt.Errorf("Player %s already exists in room", playerAddr)
	}

	room.aidBitmap = util.SetAid(room.aidBitmap, aid)
	player, err := NewPlayer(playerAddr, room, aid)
	if err != nil {
		return fmt.Errorf("Failed to create player %s due to", playerAddr)
	}

	room.players[playerAddr] = player

	if settings.GetPacketType() == settings.CombinedRace {
		go player.writeLoop(room.conn)
	}

	logging.Log("Player %s added to room", playerAddr)
	return nil
}

func RemovePlayerFromRoom(playerAddr string) error {
	p, _ := room.players[playerAddr]

	if p == nil {
		return fmt.Errorf("Player %s not in room, can't remove", playerAddr)
	}

	room.aidBitmap = util.ClearAid(room.aidBitmap, p.aid)

	delete(room.players, playerAddr)
	return nil
}

// gets the player if the player is the room's host
func GetHost(playerAddr string) *Player {
	p, _ := room.players[playerAddr]

	if p == nil {
		logging.Log("Player %s not in room, can't remove")
		return nil
	}

	return nil
}

func GetRoomAddr() string {
	return room.addr.String()
}

func GetCurrentPlayerCount() int {
	return len(room.players)
}

func CloseRoom() {
	room.conn.Close()
}

func RoomInitialized() bool {
	return room != nil
}
