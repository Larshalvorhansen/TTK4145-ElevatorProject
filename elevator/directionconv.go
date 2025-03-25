package elevator

import (
	"elevator-project/hardware"
	"fmt"
)

// Direction represents the movement direction of the elevator.
type Direction int

const (
	Up Direction = iota
	Down
)

// Converts Direction to hardware.MotorDirection
func (d Direction) ToMotorDirection() hardware.MotorDirection {
	switch d {
	case Up:
		return hardware.MD_Up
	case Down:
		return hardware.MD_Down
	default:
		fmt.Println("Warning: Invalid direction in ToMotorDirection(), returning MD_Stop")
		return hardware.MD_Stop
	}
}

// Converts Direction to hardware.ButtonType (used for button events)
func (d Direction) ToButtonType() hardware.ButtonType {
	switch d {
	case Up:
		return hardware.BT_HallUp
	case Down:
		return hardware.BT_HallDown
	default:
		fmt.Println("Warning: Invalid direction in ToButtonType(), returning invalid value")
		return -1 // Invalid ButtonType
	}
}

// Converts Direction to a readable string
func (d Direction) ToString() string {
	switch d {
	case Up:
		return "up"
	case Down:
		return "down"
	default:
		fmt.Println("Warning: Invalid direction in ToString(), returning 'unknown'")
		return "unknown"
	}
}

// Returns the opposite direction
func (d Direction) FlipDirection() Direction {
	if d != Up && d != Down {
		fmt.Println("Warning: Invalid direction in Opposite(), returning Up as default")
		return Up
	}
	return Direction(1 - d)
}
