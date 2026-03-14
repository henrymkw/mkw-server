package talker

import (
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
	OpenFroom     = 0x00
	JoinFroom	  = 0x01
	LeaveFroom    = 0x02
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

	SendToWFC(pb.Buf)
}

func HandleWFCPacket(msg []byte) {
	logging.Log("Received WFC Packet: %v", msg)

	// First byte indicates request type
	requestType := msg[0]
	switch requestType {
	case JoinFroom:
		// Actual data is after first 4 bytes
		logging.Log("Handling OpenFroom")
		joinFroomMessage := unpackJoinFroomMessage(msg[1:])
		if joinFroomMessage == nil {
			logging.Log("JoinFroomMessage is nil")
			return
		}
		handleJoinFroomMessage(joinFroomMessage)
	case LeaveFroom:
		leaveFroomMessage := unpackLeaveFroomMessage(msg[1:])
		if leaveFroomMessage == nil {
			logging.Log("Leave froom message is nil!")
			return
		}
		handleLeaveRoomRequest(leaveFroomMessage)
	default:
		logging.Log("Unknown WFC request type: %d", requestType)
	}
}

func SendToWFC(msg []byte) error {
	_, err := wfcTalker.conn.Write(msg)
	return err
}

func SendMessageToWFC(message string) error {
	if wfcTalker == nil {
		return nil
	}

	if wfcTalker.conn == nil {
		return nil
	}
	wfcTalker.conn.Write([]byte(message))
	logging.Log("Sent to WFC: %s", message)
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
