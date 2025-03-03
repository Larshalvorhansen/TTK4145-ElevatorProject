package main

import (
	"Driver-go/config"
	"Driver-go/elevator"
	"Driver-go/elevio"
	// "Driver-go/distributor"
)

func main() {
	// elevator.RunSingleElevator(222)

	newOrderC := make(chan elevator.Orders, config.Buffer)
	deliveredOrderC := make(chan elevio.ButtonEvent, config.Buffer)
	newStateC := make(chan elevator.State, config.Buffer)

	go elevator.Elevator(newOrderC, deliveredOrderC, newStateC)
}
