package main

import (
	"Driver-go/assigner"
	"Driver-go/config"
	"Driver-go/coordinator"
	"Driver-go/elevator"
	"Driver-go/hardware"
	"Driver-go/lamp"
	"Driver-go/network/bcast"
	"Driver-go/network/peers"
	"flag"
	"fmt"
	"strconv"
)

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
	confirmedCommonStateC := make(chan coordinator.SharedState, config.BufferSize)
	networkTxC := make(chan coordinator.SharedState, config.BufferSize)
	networkRxC := make(chan coordinator.SharedState, config.BufferSize)
	peersRxC := make(chan peers.PeerUpdate, config.BufferSize)
	peersTxC := make(chan bool, config.BufferSize)

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

	go coordinator.Distributor(
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
