package talker

import (
	"encoding/binary"

	"mkw-server/core"
	"mkw-server/logging"
	"mkw-server/util"
)

// Id: JoinFroom (0x01)
type JoinFroomMessage struct {
	addr     uint32
	port     uint16
	aid      uint8
	isHost   bool
	searchId uint64
}

// Id: JoinFroom (0x01)
type JoinFroomResp struct {
	searchId uint64
}

func unpackJoinFroomMessage(msg []byte) *JoinFroomMessage {
	if len(msg) != 16 {
		return nil
	}

	logging.Log("unpackNewPlayer: msg", msg)

	return &JoinFroomMessage{
		addr:     binary.BigEndian.Uint32(msg[0:4]),
		port:     binary.BigEndian.Uint16(msg[4:6]),
		aid:      uint8(msg[6]),
		isHost:   msg[7] == 1,
		searchId: binary.BigEndian.Uint64(msg[8:]),
	}
}

func packResponce(searchId uint64) []byte {
	b := make([]byte, 9)

	b[0] = JoinFroom
	binary.BigEndian.PutUint64(b[1:], searchId)
	return b
}

// addr is the address of the client that wants to join the room
func handleJoinFroomMessage(newPlayerMsg *JoinFroomMessage) {
	logging.Log("Handling WFC Join Friend Request")

	logging.Log("newPlayerMsg fields: addr:", newPlayerMsg.addr, "port", newPlayerMsg.port, "aid", newPlayerMsg.aid, "isHost", newPlayerMsg.isHost, "searchId", newPlayerMsg.searchId)

	if !core.RoomInitialized() {
		logging.Log("ERROR! Room pointer is nil, cannot handle join friend request, this should not happen!")
		return
	}

	if newPlayerMsg == nil {
		logging.Log("newPlayerMsg is nil!")
		return
	}

	addr := util.CreateUDPAddr(newPlayerMsg.addr, newPlayerMsg.port)
	if addr == nil {
		logging.Log("addr is nil in HandleJoinFroom")
		return
	}

	addPlayerResult := core.AddPlayerToRoom(addr.String())
	if !addPlayerResult {
		logging.Log("Failed to add player from WFC Join Friend Request")
		return
	}
	logging.Log("Successfully added player (", addr.String(), ") ", "to room! Current player count is", core.GetCurrentPlayerCount())

	err := SendToWFC(packResponce(newPlayerMsg.searchId))
	if err != nil {
		logging.Log("Failed to notify WFC of new player: %v", err)
		return
	}
	logging.Log("Notified WFC of new player from Join Friend Request!")
}
