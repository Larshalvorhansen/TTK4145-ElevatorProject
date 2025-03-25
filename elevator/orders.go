package elevator

import (
	"elevator-project/config"
	"elevator-project/hardware"
)

type Orders [config.NumFloors][config.NumButtons]bool

func (orders Orders) OrderInDirection(currentFloor int, direction Direction) bool {
	var start, end int
	switch direction {
	case Up:
		start, end = currentFloor+1, config.NumFloors
	case Down:
		start, end = 0, currentFloor
	default:
		panic("OrderInDirection: unknown elevator direction provided")
	}

	for floor := start; floor < end; floor++ {
		for buttonType := 0; buttonType < config.NumButtons; buttonType++ {
			if orders[floor][buttonType] {
				return true
			}
		}
	}
	return false
}

func SendCompletedOrders(
	currentFloor int,
	direction Direction,
	orders Orders,
	orderDeliveredCh chan<- hardware.ButtonEvent,
) {
	if orders[currentFloor][hardware.BT_Cab] {
		orderDeliveredCh <- hardware.ButtonEvent{Floor: currentFloor, Button: hardware.BT_Cab}
	}
	if orders[currentFloor][direction] {
		orderDeliveredCh <- hardware.ButtonEvent{Floor: currentFloor, Button: direction.ToButtonType()}
	}
}
