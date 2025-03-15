package lamp

import (
	"Driver-go/config"
	"Driver-go/distributor"
	"Driver-go/elevio"
)

func SetLamps(commonState distributor.CommonState, elevatorID int) {
	for floor := 0; floor < config.NumFloors; floor++ {
		for btn := 0; btn < 2; btn++ {
			elevio.SetButtonLamp(elevio.ButtonType(btn), floor, commonState.HallRequests[floor][btn])
		}
	}
	for floor := 0; floor < config.NumFloors; floor++ {
		elevio.SetButtonLamp(elevio.BT_Cab, floor, commonState.States[elevatorID].CabRequests[floor])
	}
}
