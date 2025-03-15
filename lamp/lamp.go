package lamp

import (
	"Driver-go/config"
	"Driver-go/distributor"
	"Driver-go/hardware"
)

func SetLamps(commonState distributor.CommonState, elevatorID int) {
	for floor := 0; floor < config.NumFloors; floor++ {
		for btn := 0; btn < 2; btn++ {
			hardware.SetButtonLamp(hardware.ButtonType(btn), floor, commonState.HallRequests[floor][btn])
		}
	}
	for floor := 0; floor < config.NumFloors; floor++ {
		hardware.SetButtonLamp(hardware.BT_Cab, floor, commonState.States[elevatorID].CabRequests[floor])
	}
}
