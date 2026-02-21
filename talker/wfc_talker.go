package talker

import (
	"net"

	"mkw-server/core"
	"mkw-server/logging"
)

// WFCTalker talks over TCP to wfc-server, mainly handling adding/removing players from the room
// its messy currently since it does a lot and the messages aren't the best (not finalized)
type WFCTalker struct {
	conn net.Conn // TCP connection to wfc-server
}

var wfcTalker *WFCTalker

const (
	RoomOpened       = 0x1 // Server responce to notify WFC that room is open
	ClientJoinFroom  = 0x2 // Used for both client requests and server responces
	ClientLeaveFroom = 0x3 // Used for both client requests and server responces
)

func NewWFCTalker(serverAddress string) error {
	conn, err := net.Dial("tcp", serverAddress)
	if err != nil {
		logging.Log("Failed to connect to WFC server: %v", err)
		return err
	}

	if core.RoomInstance == nil {
		logging.Log("Room pointer is nil when creating WFC Talker")
		return err
	}

	wfcTalker = &WFCTalker{
		conn: conn,
	}
	// immediately tell wfc-server the room address
	message := "ROOM_OPEN "
	message += core.GetRoomAddr()

	SendMessageToWFC(message)

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
			data := string(buf[:n])
			HandleWFCPacket(data)
		}
	}()
}

func HandleWFCPacket(data string) {
	logging.Log("Received WFC Packet: %v", []byte(data))

	// First byte indicates request type
	requestType := data[0]
	switch requestType {
	case ClientJoinFroom:
		// Actual data is after first 4 bytes
		HandleJoinRoomRequest(data[4:])
	case ClientLeaveFroom:
		HandleLeaveRoomRequest(data[4:])
	default:
		logging.Log("Unknown WFC request type: %d", requestType)
	}
}

// addr is the address of the client that wants to join the room
func HandleJoinRoomRequest(newClientAddr string) {
	logging.Log("Handling WFC Join Friend Request with data: %s", newClientAddr)

	if core.RoomInstance == nil {
		logging.Log("ERROR! Room pointer is nil, cannot handle join friend request, this should not happen!")
		return
	}

	addPlayerResult := core.AddPlayerToRoom(newClientAddr)
	if !addPlayerResult {
		logging.Log("Failed to add player from WFC Join Friend Request")
		return
	}
	logging.Log("Successfully added player (", newClientAddr, ") ", "to room! Current player count is", core.GetCurrentPlayerCount())

	// Responce to wfc-server is "NEW_PLAYER {player_address} {room_address}"
	// Player address is needed since it doesn't know which player sent the request
	// Room address might not be needed, maybe could be useful if mkw-server somehow gets it wrong
	message := "NEW_PLAYER "
	message += newClientAddr
	message += " "
	message += core.GetRoomAddr()

	err := SendMessageToWFC(message)
	if err != nil {
		logging.Log("Failed to notify WFC of new player: %v", err)
		return
	}
	logging.Log("Notified WFC of new player from Join Friend Request!")
}

func HandleLeaveRoomRequest(clientAddr string) {
	logging.Log("Handling WFC Leave Friend Request with data: %s", clientAddr)

	if core.RoomInstance == nil {
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

func SendMessageToWFC(message string) error {
	_, err := wfcTalker.conn.Write([]byte(message))
	logging.Log("Sent to WFC: %s", message)
	return err
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
