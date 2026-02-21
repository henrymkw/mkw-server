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
	room *core.Room // the room this controller is managing
}

// New creates a new controller with the given room address and WFC address
func New(roomAddress string, wfcAddress string) (*Controller, error) {
	room := core.NewRoom(roomAddress)
	if room == nil {
		logging.Log("Error creating room")
		return nil, errors.New("Error creating room")
	}

	err := talker.NewWFCTalker(wfcAddress, room)
	if err != nil {
		logging.Log("Error creating WFC talker: %v", err)
		return nil, err
	}

	return &Controller{
		room: room,
	}, nil
}

// Start starts the room and the WFC talker
func (c *Controller) Start() {
	talker.Start()
	c.room.Start()
}

// Close shuts down the room and the WFC talker
func (c *Controller) Close() {
	c.room.Close()
	talker.Close()
}

// NotifyShutdown notifies the WFC talker that the room is shutting down
func (c *Controller) NotifyShutdown() {
	talker.NotifyMKWServerShutdown()
}
