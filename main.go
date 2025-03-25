package main

import (
	"elevator-project/assigner"
	"elevator-project/config"
	"elevator-project/coordinator"
	"elevator-project/elevator"
	"elevator-project/hardware"
	"elevator-project/lamp"
	"elevator-project/network/bcast"
	"elevator-project/network/peers"
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

	newOrderCh := make(chan elevator.Orders, config.BufferSize)
	deliveredOrderCh := make(chan hardware.ButtonEvent, config.BufferSize)
	newStateCh := make(chan elevator.State, config.BufferSize)
	confirmedSharedStateCh := make(chan coordinator.SharedState, config.BufferSize)
	networkTxCh := make(chan coordinator.SharedState, config.BufferSize)
	networkRxCh := make(chan coordinator.SharedState, config.BufferSize)
	peersRxCh := make(chan peers.PeerUpdate, config.BufferSize)
	peersTxCh := make(chan bool, config.BufferSize)

	go peers.Receiver(
		config.MessagePort,
		peersRxCh,
	)
	go peers.Transmitter(
		config.MessagePort,
		id,
		peersTxCh,
	)

	go bcast.Receiver(
		config.MessagePort,
		networkRxCh,
	)
	go bcast.Transmitter(
		config.MessagePort,
		networkTxCh,
	)

	go coordinator.Coordinator(
		confirmedSharedStateCh,
		deliveredOrderCh,
		newStateCh,
		networkTxCh,
		networkRxCh,
		peersRxCh,
		id)

	go elevator.Elevator(
		newOrderCh,
		deliveredOrderCh,
		newStateCh,
		id)

	for {
		select {
		case ss := <-confirmedSharedStateCh:
			newOrderCh <- assigner.DistributeElevatorOrders(ss, id)
			lamp.SetLamps(ss, id)

		default:
			continue
		}
	}
}
