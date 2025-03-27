package lamp

import (
	"elevator-project/config"
	"elevator-project/coordinator"
	"elevator-project/hardware"
)

func SetButtonLamps(ss coordinator.SharedState, localID int) {
	for floor := 0; floor < config.NumFloors; floor++ {
		for button := hardware.BT_HallUp; button <= hardware.BT_HallDown; button++ {
			hardware.SetButtonLamp(button, floor, ss.HallRequests[floor][button])
		}
	}
	for floor := 0; floor < config.NumFloors; floor++ {
		hardware.SetButtonLamp(hardware.BT_Cab, floor, ss.States[localID].CabRequests[floor])
	}
}
