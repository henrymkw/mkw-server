package talker

import (
	"encoding/binary"
	"errors"
	"fmt"

	"mkw-server/core"
	"mkw-server/logging"
	"mkw-server/util"
)

// Id: JoinFroom (0x01)
type JoinFroomMessage struct {
	ip       uint32
	port     uint16
	aid      uint8
	isHost   bool
	searchId uint64
}

// Id: JoinFroom (0x01)
type JoinFroomResp struct {
	searchId uint64
}

func unpackJoinFroomMessage(msg []byte) (*JoinFroomMessage, error) {
	if len(msg) != 16 {
		return nil, fmt.Errorf("Unable to unpack JoinFroomMessage! len(msg) != 16 (%d)", len(msg))
	}

	return &JoinFroomMessage{
		ip:       binary.BigEndian.Uint32(msg[0:4]),
		port:     binary.BigEndian.Uint16(msg[4:6]),
		aid:      uint8(msg[6]),
		isHost:   msg[7] == 1,
		searchId: binary.BigEndian.Uint64(msg[8:]),
	}, nil
}

func packResponce(searchId uint64) []byte {
	b := make([]byte, 9)

	b[0] = JoinFroom
	binary.BigEndian.PutUint64(b[1:], searchId)
	return b
}

// addr is the address of the client that wants to join the room
func handleJoinFroomMessage(newPlayerMsg *JoinFroomMessage) error {
	if newPlayerMsg == nil {
		return errors.New("newPlayerMsg is nil!")
	}

	if !core.RoomInitialized() {
		return errors.New("Room isn't initialized, this should not happen at this point")
	}

	addr := util.CreateUDPAddr(newPlayerMsg.ip, newPlayerMsg.port)
	if addr == nil {
		return errors.New("CreateUDPAddr returned nil")
	}

	err := core.AddPlayerToRoom(addr.String(), newPlayerMsg.aid)
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	logging.Log("Successfully added player to room! Current player count is", core.GetCurrentPlayerCount())

	err = SendToWFC(packResponce(newPlayerMsg.searchId))
	if err != nil {
		return fmt.Errorf("Failed to notify WFC of new player: %v", err)
	}
	logging.Log("Notified WFC of new player from Join Friend Request!")
	return nil
}
