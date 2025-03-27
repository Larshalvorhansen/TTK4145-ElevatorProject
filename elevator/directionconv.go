package elevator

import (
	"elevator-project/hardware"
)

type Direction int

const (
	Up Direction = iota
	Down
)

func (d Direction) ToString() string {
	switch d {
	case Up:
		return "up"
	case Down:
		return "down"
	default:
		panic("Invalid Direction in ToString()")
	}
}

func (d Direction) FlipDirection() Direction {
	switch d {
	case Up:
		return Down
	case Down:
		return Up
	default:
		panic("Invalid Direction in FlipDirection()")
	}
}

func (d Direction) ToMotorDirection() hardware.MotorDirection {
	switch d {
	case Up:
		return hardware.MD_Up
	case Down:
		return hardware.MD_Down
	default:
		panic("Invalid Direction in ToMotorDirection()")
	}
}

func (d Direction) ToButtonType() hardware.ButtonType {
	switch d {
	case Up:
		return hardware.BT_HallUp
	case Down:
		return hardware.BT_HallDown
	default:
		panic("Invalid Direction in ToButtonType()")
	}
}
