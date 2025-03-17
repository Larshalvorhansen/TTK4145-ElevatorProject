package elevator

import (
	"Driver-go/config"
	"Driver-go/hardware"
	"fmt"
	"time"
)

type State struct {
	Obstructed  bool
	Motorstatus bool
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

func (b Behaviour) ToString() string {
	return map[Behaviour]string{
		Idle:     "idle",
		DoorOpen: "doorOpen",
		Moving:   "moving",
	}[b]
}

func Elevator(
	newOrderCh <-chan Orders,
	deliveredOrderCh chan<- hardware.ButtonEvent,
	newElevStateCh chan<- State,
) {
	doorOpenCh := make(chan bool, 16)
	doorClosedCh := make(chan bool, 16)
	floorEnteredCh := make(chan int)
	obstructedCh := make(chan bool, 16)
	motorCh := make(chan bool, 16)

	go Door(doorClosedCh, doorOpenCh, obstructedCh)
	go hardware.PollFloorSensor(floorEnteredCh)

	hardware.SetMotorDirection(hardware.MD_Down)
	state := State{Direction: Down, Behaviour: Moving}

	var orders Orders

	motorTimer := time.NewTimer(config.WatchdogTime)
	motorTimer.Stop()

	for {
		select {

		// 1) New order incoming
		case orders = <-newOrderCh:
			switch state.Behaviour {
			case Idle:
				switch {
				case orders[state.Floor][state.Direction] || orders[state.Floor][hardware.BT_Cab]:
					doorOpenCh <- true
					SendOrderDone(state.Floor, state.Direction, orders, deliveredOrderCh)
					state.Behaviour = DoorOpen
					newElevStateCh <- state

				case orders[state.Floor][state.Direction.FlipDirection()]:
					doorOpenCh <- true
					state.Direction = state.Direction.FlipDirection()
					SendOrderDone(state.Floor, state.Direction, orders, deliveredOrderCh)
					state.Behaviour = DoorOpen
					newElevStateCh <- state

				case orders.OrderInDirection(state.Floor, state.Direction):
					hardware.SetMotorDirection(state.Direction.ToMotorDirection())
					state.Behaviour = Moving
					newElevStateCh <- state
					motorTimer = time.NewTimer(config.WatchdogTime)
					motorCh <- false

				case orders.OrderInDirection(state.Floor, state.Direction.FlipDirection()):
					state.Direction = state.Direction.FlipDirection()
					hardware.SetMotorDirection(state.Direction.ToMotorDirection())
					state.Behaviour = Moving
					newElevStateCh <- state
					motorTimer = time.NewTimer(config.WatchdogTime)
					motorCh <- false
				default:
				}

			case DoorOpen:
				switch {
				case orders[state.Floor][hardware.BT_Cab] || orders[state.Floor][state.Direction]:
					doorOpenCh <- true
					SendOrderDone(state.Floor, state.Direction, orders, deliveredOrderCh)
				}

			case Moving:

			default:
				panic("Orders in wrong state")
			}

		// 2) Door closes
		case <-doorClosedCh:
			switch state.Behaviour {
			case DoorOpen:
				switch {
				case orders.OrderInDirection(state.Floor, state.Direction):
					hardware.SetMotorDirection(state.Direction.ToMotorDirection())
					state.Behaviour = Moving
					motorTimer = time.NewTimer(config.WatchdogTime)
					motorCh <- false
					newElevStateCh <- state

				case orders[state.Floor][state.Direction.FlipDirection()]:
					doorOpenCh <- true
					state.Direction = state.Direction.FlipDirection()
					SendOrderDone(state.Floor, state.Direction, orders, deliveredOrderCh)
					newElevStateCh <- state

				case orders.OrderInDirection(state.Floor, state.Direction.FlipDirection()):
					state.Direction = state.Direction.FlipDirection()
					hardware.SetMotorDirection(state.Direction.ToMotorDirection())
					state.Behaviour = Moving
					motorTimer = time.NewTimer(config.WatchdogTime)
					motorCh <- false
					newElevStateCh <- state

				default:
					state.Behaviour = Idle
					newElevStateCh <- state
				}

			default:
				panic("DoorClosed in wrong state")
			}

		// 3) Elevator finds new floor
		case state.Floor = <-floorEnteredCh:
			hardware.SetFloorIndicator(state.Floor)
			motorTimer.Stop()
			motorCh <- false

			switch state.Behaviour {
			case Moving:
				switch {
				case orders[state.Floor][state.Direction]:
					hardware.SetMotorDirection(hardware.MD_Stop)
					doorOpenCh <- true
					SendOrderDone(state.Floor, state.Direction, orders, deliveredOrderCh)
					state.Behaviour = DoorOpen

				case orders[state.Floor][hardware.BT_Cab] && orders.OrderInDirection(state.Floor, state.Direction):
					hardware.SetMotorDirection(hardware.MD_Stop)
					doorOpenCh <- true
					SendOrderDone(state.Floor, state.Direction, orders, deliveredOrderCh)
					state.Behaviour = DoorOpen

				case orders[state.Floor][hardware.BT_Cab] && !orders[state.Floor][state.Direction.FlipDirection()]:
					hardware.SetMotorDirection(hardware.MD_Stop)
					doorOpenCh <- true
					SendOrderDone(state.Floor, state.Direction, orders, deliveredOrderCh)
					state.Behaviour = DoorOpen

				case orders.OrderInDirection(state.Floor, state.Direction):
					motorTimer = time.NewTimer(config.WatchdogTime)
					motorCh <- false

				case orders[state.Floor][state.Direction.FlipDirection()]:
					hardware.SetMotorDirection(hardware.MD_Stop)
					doorOpenCh <- true
					state.Direction = state.Direction.FlipDirection()
					SendOrderDone(state.Floor, state.Direction, orders, deliveredOrderCh)
					state.Behaviour = DoorOpen

				case orders.OrderInDirection(state.Floor, state.Direction.FlipDirection()):
					state.Direction = state.Direction.FlipDirection()
					hardware.SetMotorDirection(state.Direction.ToMotorDirection())
					motorTimer = time.NewTimer(config.WatchdogTime)
					motorCh <- false

				default:
					hardware.SetMotorDirection(hardware.MD_Stop)
					state.Behaviour = Idle
				}

			default:
				panic("FloorEntered in wrong state")
			}
			newElevStateCh <- state

		// 4) MOTOR‐WATCHDOG time gone out
		case <-motorTimer.C:
			if !state.Motorstatus {
				fmt.Println("Lost motor power")
				state.Motorstatus = true
				newElevStateCh <- state
			}

		// 5) Obstruction
		case obstruction := <-obstructedCh:
			if obstruction != state.Obstructed {
				state.Obstructed = obstruction
				newElevStateCh <- state
			}

		// 6) Motor reinitialized
		case motor := <-motorCh:
			if state.Motorstatus {
				fmt.Println("Regained motor power")
				state.Motorstatus = motor
				newElevStateCh <- state
			}
		}
	}
}
