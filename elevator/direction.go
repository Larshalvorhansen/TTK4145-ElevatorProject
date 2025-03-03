// TODO: gå over funksjonene, f.eks. er Opposite() nødvendig?

package elevator

import (
	"Driver-go/elevio"
)

type Direction int

const (
	Up Direction = iota
	Down
)

// Predefined maps to avoid recreating them on every function call
var directionToMotor = map[Direction]elevio.MotorDirection{
	Up:   elevio.MD_Up,
	Down: elevio.MD_Down,
}

var directionToButton = map[Direction]elevio.ButtonType{
	Up:   elevio.BT_HallUp,
	Down: elevio.BT_HallDown,
}

var oppositeDirection = map[Direction]Direction{
	Up:   Down,
	Down: Up,
}

// Converts Direction to elevio.MotorDirection
func (d Direction) ToMotorDirection() elevio.MotorDirection {
	return directionToMotor[d]
}

// Converts Direction to elevio.ButtonType
func (d Direction) ToButtonType() elevio.ButtonType {
	return directionToButton[d]
}

// Returns the opposite direction
func (d Direction) Opposite() Direction {
	return oppositeDirection[d]
}

// Converts Direction to a string for logging/debugging
func (d Direction) ToString() string {
	switch d {
	case Up:
		return "up"
	case Down:
		return "down"
	default:
		return "unknown"
	}
}
