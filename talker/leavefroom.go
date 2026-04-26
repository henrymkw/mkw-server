package talker

import (
	"encoding/binary"
	"errors"
	"fmt"

	"mkw-server/core"
	"mkw-server/logging"
	"mkw-server/util"
)

// Id: LeaveFroom (0x02)
type LeaveFroomMessage struct {
	ip   uint32
	port uint16
}

func unpackLeaveFroomMessage(msg []byte) (*LeaveFroomMessage, error) {
	if len(msg) != 6 {
		return nil, fmt.Errorf("LeaveFroomMessage isn't 6 bytes (%d)", len(msg))
	}

	return &LeaveFroomMessage{
		ip:   binary.BigEndian.Uint32(msg[0:4]),
		port: binary.BigEndian.Uint16(msg[4:6]),
	}, nil
}

func handleLeaveRoomRequest(leaveMessage *LeaveFroomMessage) error {
	if leaveMessage == nil {
		return errors.New("LeaveMessage is nil")
	}

	logging.Log("Handling LeaveFroom. Attempting to remove player %s", util.FormatIPPort(leaveMessage.ip, leaveMessage.port))

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
