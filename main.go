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

	localPort := *serverPort
	localID := *elevatorId

	hardware.Init("localhost:"+strconv.Itoa(localPort), config.NumFloors)

	fmt.Printf("Elevator system started successfully!\n  Elevator Details:\n\tID:   %d\n\tPort: %d\n  System Configuration:\n\tFloors:    %d\n\tElevators: %d\n\n", localID, localPort, config.NumFloors, config.NumElevators)

	newOrderCh := make(chan elevator.Orders, config.BufferSize)
	orderDeliveredCh := make(chan hardware.ButtonEvent, config.BufferSize)
	localStateCh := make(chan elevator.State, config.BufferSize)
	confirmedSharedStateCh := make(chan coordinator.SharedState, config.BufferSize)
	sharedStateTxCh := make(chan coordinator.SharedState, config.BufferSize)
	sharedStateRxCh := make(chan coordinator.SharedState, config.BufferSize)
	peerUpdateRxCh := make(chan peers.PeerUpdate, config.BufferSize)
	peerEnableTxCh := make(chan bool, config.BufferSize)

	go peers.Receiver(
		config.MessagePort,
		peerUpdateRxCh,
	)
	go peers.Transmitter(
		config.MessagePort,
		localID,
		peerEnableTxCh,
	)

	go bcast.Receiver(
		config.MessagePort,
		sharedStateRxCh,
	)
	go bcast.Transmitter(
		config.MessagePort,
		sharedStateTxCh,
	)

	go coordinator.Coordinator(
		confirmedSharedStateCh,
		orderDeliveredCh,
		localStateCh,
		sharedStateTxCh,
		sharedStateRxCh,
		peerUpdateRxCh,
		localID)

	go elevator.Elevator(
		newOrderCh,
		orderDeliveredCh,
		localStateCh,
		localID)

	for {
		select {
		case ss := <-confirmedSharedStateCh:
			newOrderCh <- assigner.DistributeElevatorOrders(ss, localID)
			lamp.SetLamps(ss, localID)

		default:
			continue
		}
	}
}
