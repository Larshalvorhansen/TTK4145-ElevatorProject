package elevator

import (
	"Driver-go/config"
	"Driver-go/elevio"
	"fmt"
	"time"
)

type State struct {
	Obstructed bool
	Motorstop  bool
	Behaviour  Behaviour
	Floor      int
	Direction  Direction
}

type Behaviour int

const (
	Idle Behaviour = iota
	DoorOpen
	Moving
)

func (b Behaviour) ToString() string {
	return map[Behaviour]string{Idle: "idle", DoorOpen: "doorOpen", Moving: "moving"}[b]
}

func Elevator(
	newOrderC <-chan Orders,
	deliveredOrderC chan<- elevio.ButtonEvent,
	newStateC chan<- State,
) {

	doorOpenC := make(chan bool, 16)
	doorClosedC := make(chan bool, 16)
	floorEnteredC := make(chan int)
	obstructedC := make(chan bool, 16)
	motorC := make(chan bool, 16)

	go Door(doorClosedC, doorOpenC, obstructedC)
	go elevio.PollFloorSensor(floorEnteredC)

	elevio.SetMotorDirection(elevio.MD_Down)
	state := State{Direction: Down, Behaviour: Moving}

	var orders Orders

	motorTimer := time.NewTimer(config.WatchdogTime)
	motorTimer.Stop()

	for {
		select {
		case <-doorClosedC:
			switch state.Behaviour {
			case DoorOpen:
				switch {
				case orders.OrderInDirection(state.Floor, state.Direction):
					elevio.SetMotorDirection(state.Direction.ToMotorDirection())
					state.Behaviour = Moving
					motorTimer = time.NewTimer(config.WatchdogTime)
					motorC <- false
					newStateC <- state

				case orders[state.Floor][state.Direction.Opposite()]:
					doorOpenC <- true
					state.Direction = state.Direction.Opposite()
					OrderDone(state.Floor, state.Direction, orders, deliveredOrderC)
					newStateC <- state

				case orders.OrderInDirection(state.Floor, state.Direction.Opposite()):
					state.Direction = state.Direction.Opposite()
					elevio.SetMotorDirection(state.Direction.ToMotorDirection())
					state.Behaviour = Moving
					motorTimer = time.NewTimer(config.WatchdogTime)
					motorC <- false
					newStateC <- state

				default:
					state.Behaviour = Idle
					newStateC <- state
				}
			default:
				panic("DoorClosed in wrong state")
			}

		case state.Floor = <-floorEnteredC:
			elevio.SetFloorIndicator(state.Floor)
			motorTimer.Stop()
			motorC <- false
			switch state.Behaviour {
			case Moving:
				switch {
				case orders[state.Floor][state.Direction]:
					elevio.SetMotorDirection(elevio.MD_Stop)
					doorOpenC <- true
					OrderDone(state.Floor, state.Direction, orders, deliveredOrderC)
					state.Behaviour = DoorOpen

				case orders[state.Floor][elevio.BT_Cab] && orders.OrderInDirection(state.Floor, state.Direction):
					elevio.SetMotorDirection(elevio.MD_Stop)
					doorOpenC <- true
					OrderDone(state.Floor, state.Direction, orders, deliveredOrderC)
					state.Behaviour = DoorOpen

				case orders[state.Floor][elevio.BT_Cab] && !orders[state.Floor][state.Direction.Opposite()]:
					elevio.SetMotorDirection(elevio.MD_Stop)
					doorOpenC <- true
					OrderDone(state.Floor, state.Direction, orders, deliveredOrderC)
					state.Behaviour = DoorOpen

				case orders.OrderInDirection(state.Floor, state.Direction):
					motorTimer = time.NewTimer(config.WatchdogTime)
					motorC <- false

				case orders[state.Floor][state.Direction.Opposite()]:
					elevio.SetMotorDirection(elevio.MD_Stop)
					doorOpenC <- true
					state.Direction = state.Direction.Opposite()
					OrderDone(state.Floor, state.Direction, orders, deliveredOrderC)
					state.Behaviour = DoorOpen

				case orders.OrderInDirection(state.Floor, state.Direction.Opposite()):
					state.Direction = state.Direction.Opposite()
					elevio.SetMotorDirection(state.Direction.ToMotorDirection())
					motorTimer = time.NewTimer(config.WatchdogTime)
					motorC <- false

				default:
					elevio.SetMotorDirection(elevio.MD_Stop)
					state.Behaviour = Idle
				}
			default:
				panic("FloorEntered in wrong state")
			}
			newStateC <- state

		case orders = <-newOrderC:
			switch state.Behaviour {
			case Idle:
				switch {
				case orders[state.Floor][state.Direction] || orders[state.Floor][elevio.BT_Cab]:
					doorOpenC <- true
					OrderDone(state.Floor, state.Direction, orders, deliveredOrderC)
					state.Behaviour = DoorOpen
					newStateC <- state

				case orders[state.Floor][state.Direction.Opposite()]:
					doorOpenC <- true
					state.Direction = state.Direction.Opposite()
					OrderDone(state.Floor, state.Direction, orders, deliveredOrderC)
					state.Behaviour = DoorOpen
					newStateC <- state

				case orders.OrderInDirection(state.Floor, state.Direction):
					elevio.SetMotorDirection(state.Direction.ToMotorDirection())
					state.Behaviour = Moving
					newStateC <- state
					motorTimer = time.NewTimer(config.WatchdogTime)
					motorC <- false

				case orders.OrderInDirection(state.Floor, state.Direction.Opposite()):
					state.Direction = state.Direction.Opposite()
					elevio.SetMotorDirection(state.Direction.ToMotorDirection())
					state.Behaviour = Moving
					newStateC <- state
					motorTimer = time.NewTimer(config.WatchdogTime)
					motorC <- false
				default:
				}

			case DoorOpen:
				switch {
				case orders[state.Floor][elevio.BT_Cab] || orders[state.Floor][state.Direction]:
					doorOpenC <- true
					OrderDone(state.Floor, state.Direction, orders, deliveredOrderC)

				}

			case Moving:

			default:
				panic("Orders in wrong state")
			}
		case <-motorTimer.C:
			if !state.Motorstop {
				fmt.Println("Lost motor power")
				state.Motorstop = true
				newStateC <- state
			}
		case obstruction := <-obstructedC:
			if obstruction != state.Obstructed {
				state.Obstructed = obstruction
				newStateC <- state
			}
		case motor := <-motorC:
			if state.Motorstop {
				fmt.Println("Regained motor power")
				state.Motorstop = motor
				newStateC <- state
			}
		}
	}
}
