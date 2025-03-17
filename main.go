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
	newElevStateCh := make(chan elevator.State, config.BufferSize)
	confirmedSVCh := make(chan distributor.SystemView, config.BufferSize)
	networkTxSVCh := make(chan distributor.SystemView, config.BufferSize)
	networkRxSVCh := make(chan distributor.SystemView, config.BufferSize)
	peersRxCh := make(chan peers.PeerUpdate, config.BufferSize)
	peersTxCh := make(chan bool, config.BufferSize)

	go peers.Receiver(
		config.PeersPortNumber,
		peersRxCh,
	)
	go peers.Transmitter(
		config.PeersPortNumber,
		id,
		peersTxCh,
	)

	go bcast.Receiver(
		config.BcastPortNumber,
		networkRxSVCh,
	)
	go bcast.Transmitter(
		config.BcastPortNumber,
		networkTxSVCh,
	)

	go distributor.Distributor(
		confirmedSVCh,
		deliveredOrderCh,
		newElevStateCh,
		networkTxSVCh,
		networkRxSVCh,
		peersRxCh,
		id)

	go elevator.Elevator(
		newOrderCh,
		deliveredOrderCh,
		newElevStateCh)

	for {
		select {
		case systemView := <-confirmedSVCh:
			newOrderCh <- assigner.CalculateOptimalOrders(systemView, id)
			lamp.SetLamps(systemView, id)

		default:
			continue
		}
	}
}
