package elevator

import (
	"Driver-go/config"
	//"Driver-go/distributor"
	"Driver-go/elevio"
)

// SetLights sets the lights for the elevator system based on the current state.
func SetLights(cs distributor.CommonState, ElevatorID int) { // TODO: check if this is implemented in distributor
	// Set hall button lights
	for f := 0; f < config.NumFloors; f++ {
		for b := 0; b < 2; b++ { // Iterate over button types (up and down)
			// Set the hall button light for each floor and button type
			elevio.SetButtonLamp(elevio.ButtonType(b), f, cs.HallRequests[f][b])
		}
	}

	// Set cab button lights
	for f := 0; f < config.NumFloors; f++ {
		// Set the cab button light for each floor
		elevio.SetButtonLamp(elevio.BT_Cab, f, cs.States[ElevatorID].CabRequests[f])
	}
}
