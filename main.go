package main

import (
	"Driver-go/config"
	"Driver-go/elevator"
	"Driver-go/elevio"
	"Driver-go/lights"
	"flag"
	"fmt"
	"strconv"
)

var Port int
var id int

func main() {
	// elevator.RunSingleElevator(222)

	port := flag.Int("port", 15657, "<-- Default value, override with command line argument -port=xxxxx")
	elevatorId := flag.Int("id", 0, "<-- Default value, override with command line argument -id=x")
	flag.Parse()

	Port = *port
	id = *elevatorId

	elevio.Init("localhost:"+strconv.Itoa(Port), config.NumFloors)

	fmt.Println("Elevator initialized with id", id, "on port", Port)
	fmt.Println("System has", config.NumFloors, "floors and", config.NumElevators, "elevators.")

	newOrderC := make(chan elevator.Orders, config.Buffer)
	deliveredOrderC := make(chan elevio.ButtonEvent, config.Buffer)
	newStateC := make(chan elevator.State, config.Buffer)

	go elevator.Elevator(newOrderC, deliveredOrderC, newStateC)

	for {
		select {
		case cs := <-newStateC:
			lights.SetLights(cs, id)

		default:
			continue
		}
	}
}
