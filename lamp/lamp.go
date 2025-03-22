package lamp

import (
	"Driver-go/config"
	"Driver-go/coordinator"
	"Driver-go/hardware"
)

func SetLamps(sharedState coordinator.SharedState, elevatorID int) {
	for floor := 0; floor < config.NumFloors; floor++ {
		for btn := 0; btn < 2; btn++ {
			hardware.SetButtonLamp(hardware.ButtonType(btn), floor, sharedState.HallRequests[floor][btn])
		}
	}
	for floor := 0; floor < config.NumFloors; floor++ {
		hardware.SetButtonLamp(hardware.BT_Cab, floor, sharedState.States[elevatorID].CabRequests[floor])
	}
}
