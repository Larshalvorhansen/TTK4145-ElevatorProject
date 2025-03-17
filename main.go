package main

import (
	"Driver-go/assigner"
	"Driver-go/config"
	"Driver-go/distributor"
	"Driver-go/elevator"
	"Driver-go/hardware"
	"Driver-go/lamp"
	"Driver-go/network/bcast"
	"Driver-go/network/peers"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
)

// StartNewTerminal åpner et nytt terminalvindu og kjører "main.go" på nytt
func StartNewTerminal() {
	cmd := exec.Command("gnome-terminal", "--", "go", "run", "main.go")
	err := cmd.Start()
	if err != nil {
		fmt.Println("Feil ved åpning av nytt terminalvindu:", err)
	}
}

// MonitorSigInt lytter etter SIGINT-signal (Ctrl+C) og starter et nytt terminalvindu med "main.go"
func MonitorSigInt() {
	// Lytter etter SIGINT (Ctrl+C) eller SIGTERM (hvis prosessen stopper)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Når signalet fanges, start et nytt terminalvindu
	<-sigChan
	StartNewTerminal()
}

func main() {
	serverPort := flag.Int("port", 15657, "Elevator server port (default: 15657)")
	elevatorId := flag.Int("id", 0, "Elevator ID (default: 0)")
	flag.Parse()

	port := *serverPort
	id := *elevatorId

	hardware.Init("localhost:"+strconv.Itoa(port), config.NumFloors)

	fmt.Printf("Elevator system started successfully!\n  Elevator Details:\n\tID:   %d\n\tPort: %d\n  System Configuration:\n\tFloors:    %d\n\tElevators: %d\n\n", id, port, config.NumFloors, config.NumElevators)

	newOrderC := make(chan elevator.Orders, config.BufferSize)
	deliveredOrderC := make(chan hardware.ButtonEvent, config.BufferSize)
	newStateC := make(chan elevator.State, config.BufferSize)
	confirmedCommonStateC := make(chan distributor.CommonState, config.BufferSize)
	networkTxC := make(chan distributor.CommonState, config.BufferSize)
	networkRxC := make(chan distributor.CommonState, config.BufferSize)
	peersRxC := make(chan peers.PeerUpdate, config.BufferSize)
	peersTxC := make(chan bool, config.BufferSize)

	// Start signal monitor i en egen goroutine for å håndtere avbrudd
	go MonitorSigInt()

	go peers.Receiver(
		config.PeersPortNumber,
		peersRxC,
	)
	go peers.Transmitter(
		config.PeersPortNumber,
		id,
		peersTxC,
	)

	go bcast.Receiver(
		config.BcastPortNumber,
		networkRxC,
	)
	go bcast.Transmitter(
		config.BcastPortNumber,
		networkTxC,
	)

	go distributor.Distributor(
		confirmedCommonStateC,
		deliveredOrderC,
		newStateC,
		networkTxC,
		networkRxC,
		peersRxC,
		id)

	go elevator.Elevator(
		newOrderC,
		deliveredOrderC,
		newStateC)

	for {
		select {
		case commonState := <-confirmedCommonStateC:
			newOrderC <- assigner.CalculateOptimalOrders(commonState, id)
			lamp.SetLamps(commonState, id)

		default:
			continue
		}
	}
}
