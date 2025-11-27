package controller

import (
	"errors"

	"mkw-server/core"
	"mkw-server/logging"
	"mkw-server/talker"
)

// Controller is supposed to start and stop the room and the WFC talker
// currently it just feels like a wrapper
type Controller struct {
	room      *core.Room        // the room this controller is managing
	wfcTalker *talker.WFCTalker // Does the underlying communication with WFC
}

// New creates a new controller with the given room address and WFC address
func New(roomAddress string, wfcAddress string) (*Controller, error) {
	room := core.NewRoom(roomAddress)
	if room == nil {
		logging.Log("Error creating room")
		return nil, errors.New("Error creating room")
	}

	wfcTalker, err := talker.NewWFCTalker(wfcAddress, room)
	if err != nil {
		logging.Log("Error creating WFC talker: %v", err)
		return nil, err
	}

	return &Controller{
		room:      room,
		wfcTalker: wfcTalker,
	}, nil
}

// Start starts the room and the WFC talker
func (c *Controller) Start() {
	c.wfcTalker.Start()
	c.room.Start()
}

// Close shuts down the room and the WFC talker
func (c *Controller) Close() {
	c.room.Close()
	c.wfcTalker.Close()
}

// NotifyShutdown notifies the WFC talker that the room is shutting down
func (c *Controller) NotifyShutdown() {
	c.wfcTalker.NotifyMKWServerShutdown()
}
