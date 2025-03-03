package elevator

import (
	"Driver-go/config"
	"Driver-go/elevio"
)

// Orders represents a matrix that tracks active button orders in the elevator system.
// The first index is the floor, and the second index is the button type (Up, Down, Cab).
type Orders [config.NumFloors][config.NumButtons]bool

// OrderInDirection checks if there are any active orders in the given direction from the specified floor.
// Returns true if there is an order in the specified direction, otherwise false.
func (a Orders) OrderInDirection(floor int, dir Direction) bool {
	switch dir {
	case Up:
		// Check floors above the current floor for any active orders
		for f := floor + 1; f < config.NumFloors; f++ {
			for b := 0; b < config.NumButtons; b++ {
				if a[f][b] {
					return true // Order found in the upward direction
				}
			}
		}
		return false
	case Down:
		// Check floors below the current floor for any active orders
		for f := 0; f < floor; f++ {
			for b := 0; b < config.NumButtons; b++ {
				if a[f][b] {
					return true // Order found in the downward direction
				}
			}
		}
		return false
	default:
		// If the direction is invalid, trigger a panic
		panic("Invalid direction")
	}
}

// OrderDone sends a signal when an order at the specified floor has been completed.
// This function notifies the system via the orderDoneC channel.
func OrderDone(floor int, dir Direction, a Orders, orderDoneC chan<- elevio.ButtonEvent) {
	// If there was a cab call at this floor, notify that it has been handled
	if a[floor][elevio.BT_Cab] {
		orderDoneC <- elevio.ButtonEvent{Floor: floor, Button: elevio.BT_Cab}
	}

	// If there was a hall call in the current direction, notify that it has been handled
	if a[floor][dir] {
		orderDoneC <- elevio.ButtonEvent{Floor: floor, Button: dir.ToButtonType()}
	}
}
