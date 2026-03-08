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
	OpenRoom     = 0x00
	AddPlayer    = 0x01
	CloseRoom    = 0x02
	RemovePlayer = 0x03
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

	pb.WriteUint8(OpenRoom)
	pb.WriteUint16(wfcTalker.port)

	SendToWFC(pb.Buf)
}

func HandleWFCPacket(msg []byte) {
	logging.Log("Received WFC Packet: %v", msg)

	// First byte indicates request type
	requestType := msg[0]
	switch requestType {
	case AddPlayer:
		// Actual data is after first 4 bytes
		logging.Log("Handling Add Player!")
		handleAddPlayer(unpackNewPlayer(msg[1:]))
	case RemovePlayer:
		// HandleLeaveRoomRequest(data[1:])
	default:
		logging.Log("Unknown WFC request type: %d", requestType)
	}
}

func HandleLeaveRoomRequest(clientAddr string) {
	logging.Log("Handling WFC Leave Friend Request with data: %s", clientAddr)

	if !core.RoomInitialized() {
		logging.Log("ERROR! Room pointer is nil, cannot handle leave friend request, this should not happen!")
		return
	}

	removePlayerResult := core.RemovePlayerFromRoom(clientAddr)
	if !removePlayerResult {
		logging.Log("Failed to remove player from WFC Leave Friend Request")
		return
	}

	logging.Log("Successfully removed player (", clientAddr, ") ", "from room! Current player count is", core.GetCurrentPlayerCount())
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

func NotifyMKWServerShutdown() error {
	message := string("MKWSERVER_SHUTDOWN ")
	message += core.GetRoomAddr()
	return SendMessageToWFC(message)
}

func TalkerInitialized() bool {
	return wfcTalker != nil
}
