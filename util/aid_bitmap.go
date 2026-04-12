package util

import (
	"errors"
)

func IsSet(aids uint16, aid byte) bool {
	return (aids & (1 << aid)) != 0
}

func GetSendToAids(aidBitmap uint16) *[]byte {
	aidsToSendTo := []byte{}
	var i byte
	for i = range 12 {
		if IsSet(aidBitmap, i) {
			aidsToSendTo = append(aidsToSendTo, i)
		}
	}
	return &aidsToSendTo
}

func AidSlot(aid uint8) uint32 {
	return 1 << aid
}

func SetAid(aids uint32, aid uint8) uint32 {
	return aids | AidSlot(aid)
}

func ClearAid(aids uint32, aid uint8) uint32 {
	return aids & ^AidSlot(aid)
}

func GetAvailableAid(aidBitmap uint32) (uint8, error) {
	var i uint8
	for i = range 12 {
		if ((aidBitmap >> i) & 1) == 0 {
			return i, nil
		}
	}
	return 0xff, errors.New("No available aid!")
}
