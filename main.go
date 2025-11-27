package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"mkw-server/controller"
	"mkw-server/logging"
)

// mkw-server starts when a group is created in wfc-server and a group is
// created in wfc-server when it matches two players, well before players are notified they're
// in a group. mkw-server has time to setup and start listeners before players are notified
func main() {
	logging.InitLogFile()
	defer logging.CloseLogFile()

	// roomAddr is the address players send to and receive from
	// reason this gets passed in is to help mkw-server manage multiple rooms,
	// could also be useful in the case mkw-server isnt running on the same machine as wfc-serve
	roomAddr := flag.String("room-addr", "", "UDP address (ip:port) the room listens on")

	// wfcAddr is the address of the wfc-server controller
	// This is how mkw-server and wfc-server will communicate
	wfcAddr := flag.String("wfc-addr", "", "TCP address (ip:port) of the WFC controller")

	flag.Parse()

	// Validate required arguments
	if *roomAddr == "" {
		logging.Log("Missing required argument: --room-addr {ip:port}")
		os.Exit(1)
	}
	if *wfcAddr == "" {
		logging.Log("Missing required argument: --wfc-addr {ip:port}")
		os.Exit(1)
	}

	logging.Log("Starting mkw-server. Room address=%s wfc-server address=%s\n", *roomAddr, *wfcAddr)

	controller, err := controller.New(*roomAddr, *wfcAddr)

	if err != nil {
		logging.Log("Failed to initialize server: %v", err)
		os.Exit(1)
	}

	// Start the controller (which starts the room and WFC talker)
	controller.Start()

	// mkw-server controls the lifetime of mkw-server and will send a SIGTERM or SIGINT when it wants to close
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	<-sigCh
	logging.Log("Shutting down mkw-server...")
	logging.CloseLogFile()

	controller.NotifyShutdown()

	controller.Close()
}
