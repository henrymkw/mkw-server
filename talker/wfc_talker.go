package talker

import (
	"errors"
	"net"

	"mkw-server/core"
	"mkw-server/logging"
	"mkw-server/util"
)

// WFCTalker talks over TCP to wfc-server, mainly handling adding/removing players from the room
// its messy currently since it does a lot and the messages aren't the best (not finalized)
type WFCTalker struct {
	conn net.Conn // TCP connection to wfc-server
	port uint16   // room's udp port, used as an id (a bad one)
}

var wfcTalker *WFCTalker

type MKWServerMessage uint8

const (
	OpenFroom      = 0x00
	JoinFroom      = 0x01
	LeaveFroom     = 0x02
	LogToWFCServer = 0xff
)

func NewWFCTalker(port uint16, serverAddress string) error {
	conn, err := net.Dial("tcp", serverAddress)
	if err != nil {
		logging.Log("Failed to connect to WFC server: %v", err)
		return err
	}

	if !core.RoomInitialized() {
		logging.Log("Room pointer is nil when creating WFC Talker")
		return err
	}

	wfcTalker = &WFCTalker{
		conn: conn,
		port: port,
	}
	logging.Log("Talker created! (Hello wfc-server!)")

	// immediately tell wfc-server the room address
	sendRoomOpen()

	return nil
}

func Start() {
	go func() {
		buf := make([]byte, 128)
		for {
			n, err := wfcTalker.conn.Read(buf)
			if err != nil {
				errMsg := "Error reading from WFC server"
				logging.Log(errMsg)
				return
			}
			data := buf[:n]
			HandleWFCPacket(data)
		}
	}()
}

func sendRoomOpen() {
	pb := &util.PacketBuilder{Buf: make([]byte, 0, 3)}

	pb.WriteUint8(OpenFroom)
	pb.WriteUint16(wfcTalker.port)

	logging.Log("Sending OpenRoom to wfc-server")
	SendToWFC(pb.Buf)
}

func HandleWFCPacket(msg []byte) {
	// First byte indicates request type
	requestType := msg[0]
	switch requestType {
	case JoinFroom:
		logging.Log("Received JoinFroom")
		joinFroomMessage, err := unpackJoinFroomMessage(msg[1:])
		if err != nil {
			logging.Log(err.Error())
			return
		}

		err = handleJoinFroomMessage(joinFroomMessage)
		if err != nil {
			logging.Log("Failed to handle JoinFroom for reason %s:", err.Error())
		}
		logging.Log("Successfully handled JoinFroom for player %s", util.FormatIPPort(joinFroomMessage.ip, joinFroomMessage.port))

	case LeaveFroom:
		logging.Log("Received LeaveFroom")
		leaveFroomMessage, err := unpackLeaveFroomMessage(msg[1:])
		if err != nil {
			logging.Log("Unable to unpack LeaveFroomMessage for reason %s", err.Error())
			return
		}

		err = handleLeaveRoomRequest(leaveFroomMessage)
		if err != nil {
			logging.Log("Failed to handle LeaveFroom for reason %s:", err.Error())
		}
		logging.Log("Successfully handled LeaveFroom for player %s", util.FormatIPPort(leaveFroomMessage.ip, leaveFroomMessage.port))

	default:
		logging.Log("Received Unknown Request %d from wfc-server of length %d", requestType, len(msg))
	}
}

func SendToWFC(msg []byte) error {
	// Messages need to be wrapped around delimiters to prevent
	// messages from being streamed together.

	// messages are prefixed with 0xbb, 0xef, 0xdc, 0xc8
	// and suffixed with 0xce, 0xf9, 0xd3, 0xaa

	out := append([]byte{0xbb, 0xef, 0xdc, 0xc8}, msg...)
	out = append(out, []byte{0xce, 0xf9, 0xd3, 0xaa}...)
	_, err := wfcTalker.conn.Write(out)
	return err
}

func SendMessageToWFC(message string) error {
	if wfcTalker == nil {
		return errors.New("Can't send to wfc-server, talker is nil")
	}

	if wfcTalker.conn == nil {
		return errors.New("Can't send to wfc-server, talker.conn is nil")
	}

	// wfc-server knows to log mkw-server logs with LogToWFCServer (0xff)
	data := append([]byte{LogToWFCServer}, message...)

	err := SendToWFC(data)
	if err != nil {
		return err
	}

	return nil
}

func SendPacketDataToWFC(data []byte) error {
	_, err := wfcTalker.conn.Write(data)
	if err != nil {
		logging.Log("Failed to send packet data to WFC: %v", err)
		return err
	}
	logging.Log("Sent %d bytes of packet data to WFC", len(data))
	return nil
}

func Close() {
	wfcTalker.conn.Close()
}

func TalkerInitialized() bool {
	return wfcTalker != nil
}
