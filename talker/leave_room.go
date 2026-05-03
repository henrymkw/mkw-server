package talker

import (
	"encoding/binary"
	"errors"
	"fmt"

	"mkw-server/core"
	"mkw-server/logging"
	"mkw-server/util"
)

// Id: LeaveRoom (0x02)
type LeaveRoomMessage struct {
	ip   uint32
	port uint16
}

func unpackLeaveRoomMessage(msg []byte) (*LeaveRoomMessage, error) {
	if len(msg) != 6 {
		return nil, fmt.Errorf("LeaveRoomMessage isn't 6 bytes (%d)", len(msg))
	}

	return &LeaveRoomMessage{
		ip:   binary.BigEndian.Uint32(msg[0:4]),
		port: binary.BigEndian.Uint16(msg[4:6]),
	}, nil
}

func handleLeaveRoomRequest(leaveMessage *LeaveRoomMessage) error {
	if leaveMessage == nil {
		return errors.New("LeaveMessage is nil")
	}

	logging.Log("Handling LeaveRoom. Attempting to remove player %s", util.FormatIPPort(leaveMessage.ip, leaveMessage.port))

	if !core.RoomInitialized() {
		return errors.New("Room isn't initialized, this shouldn't happen at this point")
	}

	addr := util.CreateUDPAddr(leaveMessage.ip, leaveMessage.port)
	if addr == nil {
		return errors.New("CreateUDPAddr failed in handleLeaveRoomRequest")
	}

	err := core.RemovePlayerFromRoom(addr.String())
	if err != nil {
		return err
	}

	return nil
}
