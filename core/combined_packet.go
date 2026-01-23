package core

import "encoding/binary"

// A combined packet is a custom packet format for mkw-server
// where multiple player's race packets can be combined into a single packet
//
// Note that this is experimental and its not clear if this is any better
// or worse from a lag pov than just sending one packet at a time

// Technically the header isn't needed since the Race packet's
// header contains the sizes for each record, but this
// is safer and makes parsing these packets easier for the client
type CombinedPacketHeader struct {
	Magic      uint8    // Currently 0xD
	NumPackets uint8    // Number of packets in this combined packet
	TotalSize  uint16   // Total size of combined packet including header
	Offsets    []uint16 // Offsets to each packet (relative to start of packet data). Length is NumPackets
}

// GetHeaderSize returns the size of the header for N packets
// Structure: 1 byte magic + 1 byte numPackets + 2 bytes totalSize + 2*N bytes offsets
func GetHeaderSize(numPackets int) int {
	return 4 + (2 * numPackets)
}

// BuildCombinedPacket takes a slice of packet data and builds combined packets with headers
// It splits packets across multiple combined packets if they exceed maxSize
func BuildCombinedPacket(packets [][]byte, maxSize int) [][]byte {
	if len(packets) == 0 {
		return nil
	}

	result := make([][]byte, 0)
	currentBatch := make([][]byte, 0)
	currentSize := 0

	for _, pkt := range packets {
		// Calculate size if we add this packet
		newBatchSize := len(currentBatch) + 1
		headerSize := GetHeaderSize(newBatchSize)
		totalSize := headerSize + currentSize + len(pkt)

		if totalSize > maxSize && len(currentBatch) > 0 {
			// Current batch is full, encode and start new one
			result = append(result, encodeCombinedPacket(currentBatch))
			currentBatch = [][]byte{pkt}
			currentSize = len(pkt)
		} else {
			// Add to current batch
			currentBatch = append(currentBatch, pkt)
			currentSize += len(pkt)
		}
	}

	// Encode remaining packets
	if len(currentBatch) > 0 {
		result = append(result, encodeCombinedPacket(currentBatch))
	}

	return result
}

// encodeCombinedPacket encodes a batch of packets with the combined packet header
func encodeCombinedPacket(packets [][]byte) []byte {
	if len(packets) == 0 {
		return nil
	}

	numPackets := len(packets)
	headerSize := GetHeaderSize(numPackets)

	// Calculate total data size
	dataSize := 0
	for _, pkt := range packets {
		dataSize += len(pkt)
	}

	totalSize := headerSize + dataSize
	result := make([]byte, totalSize)

	// Write fixed header (4 bytes)
	result[0] = 0xD // Magic byte
	result[1] = uint8(numPackets)
	binary.BigEndian.PutUint16(result[2:4], uint16(totalSize))

	// Write offsets array (2 bytes per packet)
	offset := 4 + (2 * numPackets) // Start of packet data
	for i, pkt := range packets {
		binary.BigEndian.PutUint16(result[4+i*2:4+i*2+2], uint16(offset))
		offset += len(pkt)
	}

	// Write packet data
	dataStart := headerSize
	for _, pkt := range packets {
		copy(result[dataStart:dataStart+len(pkt)], pkt)
		dataStart += len(pkt)
	}

	return result
}

// ParseCombinedPacketHeader parses the header from a combined packet
func ParseCombinedPacketHeader(data []byte) *CombinedPacketHeader {
	if len(data) < 4 {
		return nil
	}

	magic := uint8(data[0])
	numPackets := uint8(data[1])
	totalSize := binary.BigEndian.Uint16(data[2:4])

	headerSize := GetHeaderSize(int(numPackets))
	if len(data) < headerSize {
		return nil
	}

	offsets := make([]uint16, numPackets)
	for i := 0; i < int(numPackets); i++ {
		offsets[i] = binary.BigEndian.Uint16(data[4+i*2 : 4+i*2+2])
	}

	return &CombinedPacketHeader{
		Magic:      magic,
		NumPackets: numPackets,
		TotalSize:  totalSize,
		Offsets:    offsets,
	}
}
