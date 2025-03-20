// TODO: Check if this statemachine should be more like the one in distributor/statemachine.go
// TODO: check iota constants, if they are exported capitalize firt letter, if not, keep it lowercase in first letter
package elevator

import (
	"Driver-go/config"
	"Driver-go/hardware"
	"log"
	"time"
)

// State holds the elevator's current status.
type State struct {
	Obstructed  bool
	MotorStatus bool
	Behaviour   Behaviour
	Floor       int
	Direction   Direction
}

type Behaviour int

const (
	Idle Behaviour = iota
	DoorOpen
	Moving
)

// Helper checks if there's an order at the current floor for either the given direction or cab.
func hasOrder(orders Orders, floor int, dir Direction) bool {
	if orders[floor][dir.ToButtonType()] || orders[floor][hardware.BT_Cab] {
		return true
	}
	return false
}

// Helper starts motor movement with the given direction and resets the watchdog timer.
func startMotor(dir Direction, state *State, motorTimer *time.Timer, motorC chan<- bool, newStateC chan<- State) {
	hardware.SetMotorDirection(dir.ToMotorDirection())
	state.Behaviour = Moving
	motorTimer.Stop()
	motorTimer.Reset(config.WatchdogTime)
	motorC <- false
	newStateC <- *state
}

func (b Behaviour) ToString() string {
	return map[Behaviour]string{
		Idle:     "idle",
		DoorOpen: "doorOpen",
		Moving:   "moving",
	}[b]
}

func Elevator(
	newOrderC <-chan Orders,
	deliveredOrderC chan<- hardware.ButtonEvent,
	newStateC chan<- State,
) {
	doorOpenC := make(chan bool, config.ElevatorChannelBuf)
	doorClosedC := make(chan bool, config.ElevatorChannelBuf)
	floorEnteredC := make(chan int)
	obstructedC := make(chan bool, config.ElevatorChannelBuf)
	motorC := make(chan bool, config.ElevatorChannelBuf)

	go DoorLogic(doorClosedC, doorOpenC, obstructedC)
	go hardware.PollFloorSensor(floorEnteredC)

	hardware.SetMotorDirection(hardware.MD_Down)
	state := State{Direction: Down, Behaviour: Moving}

	var orders Orders

	motorTimer := time.NewTimer(config.WatchdogTime)
	motorTimer.Stop()

	for {
		select {

		// 1) New order
		case orders = <-newOrderC:
			switch state.Behaviour {
			case Idle:
				switch {
				case hasOrder(orders, state.Floor, state.Direction):
					doorOpenC <- true
					SendOrderDone(state.Floor, state.Direction, orders, deliveredOrderC)
					state.Behaviour = DoorOpen
					newStateC <- state

				case hasOrder(orders, state.Floor, state.Direction.FlipDirection()):
					doorOpenC <- true
					state.Direction = state.Direction.FlipDirection()
					SendOrderDone(state.Floor, state.Direction, orders, deliveredOrderC)
					state.Behaviour = DoorOpen
					newStateC <- state

				case orders.OrderInDirection(state.Floor, state.Direction):
					startMotor(state.Direction, &state, motorTimer, motorC, newStateC)

				case orders.OrderInDirection(state.Floor, state.Direction.FlipDirection()):
					state.Direction = state.Direction.FlipDirection()
					startMotor(state.Direction, &state, motorTimer, motorC, newStateC)
				}

			case DoorOpen:
				if hasOrder(orders, state.Floor, state.Direction) {
					doorOpenC <- true
					SendOrderDone(state.Floor, state.Direction, orders, deliveredOrderC)
				}

			case Moving:
				// No immediate action on new orders while moving.

			default:
				panic("Orders in wrong state")
			}

		// 2) Door closes
		case <-doorClosedC:
			switch state.Behaviour {
			case DoorOpen:
				switch {
				case orders.OrderInDirection(state.Floor, state.Direction):
					startMotor(state.Direction, &state, motorTimer, motorC, newStateC)

				case hasOrder(orders, state.Floor, state.Direction.FlipDirection()):
					doorOpenC <- true
					state.Direction = state.Direction.FlipDirection()
					SendOrderDone(state.Floor, state.Direction, orders, deliveredOrderC)
					newStateC <- state

				case orders.OrderInDirection(state.Floor, state.Direction.FlipDirection()):
					state.Direction = state.Direction.FlipDirection()
					startMotor(state.Direction, &state, motorTimer, motorC, newStateC)

				default:
					state.Behaviour = Idle
					newStateC <- state
				}
			default:
				panic("DoorClosed in wrong state")
			}

		// 3) Floor sensor
		case state.Floor = <-floorEnteredC:
			hardware.SetFloorIndicator(state.Floor)
			motorTimer.Stop()
			motorC <- false
			switch state.Behaviour {
			case Moving:
				switch {
				case hasOrder(orders, state.Floor, state.Direction):
					hardware.SetMotorDirection(hardware.MD_Stop)
					doorOpenC <- true
					SendOrderDone(state.Floor, state.Direction, orders, deliveredOrderC)
					state.Behaviour = DoorOpen

				case orders[state.Floor][hardware.BT_Cab] && orders.OrderInDirection(state.Floor, state.Direction):
					hardware.SetMotorDirection(hardware.MD_Stop)
					doorOpenC <- true
					SendOrderDone(state.Floor, state.Direction, orders, deliveredOrderC)
					state.Behaviour = DoorOpen

				case orders[state.Floor][hardware.BT_Cab] && !orders[state.Floor][state.Direction.FlipDirection()]:
					hardware.SetMotorDirection(hardware.MD_Stop)
					doorOpenC <- true
					SendOrderDone(state.Floor, state.Direction, orders, deliveredOrderC)
					state.Behaviour = DoorOpen

				case orders.OrderInDirection(state.Floor, state.Direction):
					motorTimer.Stop()
					motorTimer.Reset(config.WatchdogTime)
					motorC <- false

				case orders[state.Floor][state.Direction.FlipDirection()]:
					hardware.SetMotorDirection(hardware.MD_Stop)
					doorOpenC <- true
					state.Direction = state.Direction.FlipDirection()
					SendOrderDone(state.Floor, state.Direction, orders, deliveredOrderC)
					state.Behaviour = DoorOpen

				case orders.OrderInDirection(state.Floor, state.Direction.FlipDirection()):
					state.Direction = state.Direction.FlipDirection()
					motorTimer.Stop()
					motorTimer.Reset(config.WatchdogTime)
					motorC <- false

				default:
					hardware.SetMotorDirection(hardware.MD_Stop)
					state.Behaviour = Idle
				}
			default:
				panic("FloorEntered in wrong state")
			}
			newStateC <- state

		// 4) Motor watchdog
		case <-motorTimer.C:
			if !state.MotorStatus {
				log.Println("Lost motor power")
				state.MotorStatus = true
				newStateC <- state
			}

		// 5) Obstruction
		case obstruction := <-obstructedC:
			if obstruction != state.Obstructed {
				state.Obstructed = obstruction
				newStateC <- state
			}

		// 6) Motor reinit
		case motor := <-motorC:
			if state.MotorStatus {
				log.Println("Regained motor power")
				state.MotorStatus = motor
				newStateC <- state
			}
		}
	}
}
