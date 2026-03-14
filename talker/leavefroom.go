package talker

import (
	"encoding/binary"

	"mkw-server/core"
	"mkw-server/logging"
	"mkw-server/util"
)

// Id: LeaveFroom (0x02)
type LeaveFroomMessage struct {
	ip 			uint32
	port 		uint16
}

func unpackLeaveFroomMessage(msg []byte) *LeaveFroomMessage {
	if len(msg) != 6 {
		logging.Log("LeaveFroomMessage isn't 6 bytes (%d)", len(msg))
		return nil
	}
	
	logging.Log("")

	return &LeaveFroomMessage{
		ip:		binary.BigEndian.Uint32(msg[0:4]),
		port:		binary.BigEndian.Uint16(msg[4:6]),
	}
}

func handleLeaveRoomRequest(leaveMessage *LeaveFroomMessage) {
	logging.Log("Handling WFC Leave Friend Request")

	if !core.RoomInitialized() {
		logging.Log("ERROR! Room pointer is nil, cannot handle leave friend request, this should not happen!")
		return
	}

	logging.Log("Room is initialized")

	if leaveMessage == nil {
		logging.Log("LeaveMEssage is nil")
	}

	addr := util.CreateUDPAddr(leaveMessage.ip, leaveMessage.port)
	if addr == nil {
		logging.Log("addr is nil in handleLeaveRoomRequest")
		return
	}

	logging.Log("UDP Addr created")

	removePlayerResult := core.RemovePlayerFromRoom(addr.String())
	if !removePlayerResult {
		logging.Log("Failed to remove player from WFC Leave Friend Request")
		return
	}

	logging.Log("Successfully removed player (", addr.String(), ") ", "from room! Current player count is", core.GetCurrentPlayerCount())
}


