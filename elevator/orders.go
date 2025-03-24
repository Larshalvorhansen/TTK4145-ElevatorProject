package elevator

import (
	"Driver-go/config"
	"Driver-go/hardware"
)

type Orders [config.NumFloors][config.NumButtons]bool

func (o Orders) OrderInDirection(floor int, dir Direction) bool {
	var start, end int
	switch dir {
	case Up:
		start, end = floor+1, config.NumFloors
	case Down:
		start, end = 0, floor
	default:
		panic("Invalid direction")
	}

	for f := start; f < end; f++ {
		for b := 0; b < config.NumButtons; b++ {
			if o[f][b] {
				return true
			}
		}
	}
	return false
}

func OrderAtFloor(floor int, orders Orders) bool {
	return orders[floor][Up] || orders[floor][Down] || orders[floor][hardware.BT_Cab]
}



func SendOrderDone(
	floor int,
	dir Direction,
	orders Orders,
	orderDoneCh chan<- hardware.ButtonEvent,
) {
	if orders[floor][hardware.BT_Cab] {
		orderDoneCh <- hardware.ButtonEvent{Floor: floor, Button: hardware.BT_Cab}
	}
	if orders[floor][dir] {
		orderDoneCh <- hardware.ButtonEvent{Floor: floor, Button: dir.ToButtonType()}
	}
}
