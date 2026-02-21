package settings

// atm, this type is silly since theres only one setting that can't be changed without recompiling
// When host settings are implemented in-game this will make more sense
type Settings struct {
	packetType RacePacketType
}

type RacePacketType int

var RoomSettings Settings

// RacePacketType
const (
	Race         = 0 // one packet at a time, default value
	CombinedRace = 1 // packets are batched together if possible. set by passed in --combined
)

func InitDefaultSettings() {
	RoomSettings = Settings{
		packetType: Race,
	}
}

func GetPacketType() RacePacketType {
	return RoomSettings.packetType
}
