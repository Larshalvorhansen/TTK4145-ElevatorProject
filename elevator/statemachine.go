package elevator

import (
	"elevator-project/config"
	"elevator-project/hardware"
	"fmt"
	"strconv"
	"time"
)

// logEvent: En hjelpefunksjon for å skrive ut viktige hendelser
func logEvent(id int, format string, args ...interface{}) {
	fmt.Printf("[Elevator "+strconv.Itoa(id)+"] "+format+"\n", args...)
}

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
	orderDeliveredCh chan<- hardware.ButtonEvent,
	localStateCh chan<- State,
	id int,
) {
	doorOpenCh := make(chan bool, config.ElevatorChBuffer)
	doorClosedCh := make(chan bool, config.ElevatorChBuffer)
	floorEnteredCh := make(chan int)
	obstructedCh := make(chan bool, config.ElevatorChBuffer)
	motorCh := make(chan bool, config.ElevatorChBuffer)

	go DoorLogic(doorClosedCh, doorOpenCh, obstructedCh)
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
			// logEvent(id, "New order received") // For debugging
			switch state.Behaviour {
			case Idle:
				switch {
				case orders[state.Floor][state.Direction] || orders[state.Floor][hardware.BT_Cab]:
					doorOpenCh <- true
					SendOrderDone(state.Floor, state.Direction, orders, orderDeliveredCh)
					state.Behaviour = DoorOpen
					localStateCh <- state

				case orders[state.Floor][state.Direction.FlipDirection()]:
					doorOpenCh <- true
					state.Direction = state.Direction.FlipDirection()
					SendOrderDone(state.Floor, state.Direction, orders, orderDeliveredCh)
					state.Behaviour = DoorOpen
					localStateCh <- state

				case orders.OrderInDirection(state.Floor, state.Direction):
					hardware.SetMotorDirection(state.Direction.ToMotorDirection())
					state.Behaviour = Moving
					localStateCh <- state
					motorTimer = time.NewTimer(config.WatchdogTime)
					motorCh <- false

				case orders.OrderInDirection(state.Floor, state.Direction.FlipDirection()):
					state.Direction = state.Direction.FlipDirection()
					hardware.SetMotorDirection(state.Direction.ToMotorDirection())
					state.Behaviour = Moving
					localStateCh <- state
					motorTimer = time.NewTimer(config.WatchdogTime)
					motorCh <- false
				default:
				}

			case DoorOpen:
				switch {
				case orders[state.Floor][hardware.BT_Cab] || orders[state.Floor][state.Direction]:
					doorOpenCh <- true
					SendOrderDone(state.Floor, state.Direction, orders, orderDeliveredCh)
				}

			case Moving:

			default:
				panic("Orders in wrong state")
			}

		// 2) DoorLogic closes
		case <-doorClosedCh:
			logEvent(id, "Door closed at floor %d", state.Floor) // For debugging
			switch state.Behaviour {
			case DoorOpen:
				switch {
				case orders.OrderInDirection(state.Floor, state.Direction):
					hardware.SetMotorDirection(state.Direction.ToMotorDirection())
					state.Behaviour = Moving
					motorTimer = time.NewTimer(config.WatchdogTime)
					motorCh <- false
					localStateCh <- state

				case orders[state.Floor][state.Direction.FlipDirection()]:
					doorOpenCh <- true
					state.Direction = state.Direction.FlipDirection()
					SendOrderDone(state.Floor, state.Direction, orders, orderDeliveredCh)
					localStateCh <- state

				case orders.OrderInDirection(state.Floor, state.Direction.FlipDirection()):
					state.Direction = state.Direction.FlipDirection()
					hardware.SetMotorDirection(state.Direction.ToMotorDirection())
					state.Behaviour = Moving
					motorTimer = time.NewTimer(config.WatchdogTime)
					motorCh <- false
					localStateCh <- state

				default:
					state.Behaviour = Idle
					localStateCh <- state
				}

			default:
				panic("Door in wrong state")
			}

		// 3) Elevator finds new floor
		case state.Floor = <-floorEnteredCh:
			logEvent(id, "Detected floor %d", state.Floor) // For debugging
			hardware.SetFloorIndicator(state.Floor)
			motorTimer.Stop()
			motorCh <- false

			switch state.Behaviour {
			case Moving:
				switch {
				case orders[state.Floor][state.Direction]:
					hardware.SetMotorDirection(hardware.MD_Stop)
					doorOpenCh <- true
					SendOrderDone(state.Floor, state.Direction, orders, orderDeliveredCh)
					state.Behaviour = DoorOpen

				case orders[state.Floor][hardware.BT_Cab] && orders.OrderInDirection(state.Floor, state.Direction):
					hardware.SetMotorDirection(hardware.MD_Stop)
					doorOpenCh <- true
					SendOrderDone(state.Floor, state.Direction, orders, orderDeliveredCh)
					state.Behaviour = DoorOpen

				case orders[state.Floor][hardware.BT_Cab] && !orders[state.Floor][state.Direction.FlipDirection()]:
					hardware.SetMotorDirection(hardware.MD_Stop)
					doorOpenCh <- true
					SendOrderDone(state.Floor, state.Direction, orders, orderDeliveredCh)
					state.Behaviour = DoorOpen

				case orders.OrderInDirection(state.Floor, state.Direction):
					motorTimer = time.NewTimer(config.WatchdogTime)
					motorCh <- false

				case orders[state.Floor][state.Direction.FlipDirection()]:
					hardware.SetMotorDirection(hardware.MD_Stop)
					doorOpenCh <- true
					state.Direction = state.Direction.FlipDirection()
					SendOrderDone(state.Floor, state.Direction, orders, orderDeliveredCh)
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
			localStateCh <- state

		// 4) MOTOR‐WATCHDOG time gone out
		case <-motorTimer.C:
			if !state.Motorstatus {
				logEvent(id, "WARNING: Lost motor power!") // For debugging
				// fmt.Println("Lost motor power")
				state.Motorstatus = true
				localStateCh <- state
			}

		// 5) Obstruction
		case obstruction := <-obstructedCh:
			if obstruction != state.Obstructed {
				state.Obstructed = obstruction
				if obstruction {
					logEvent(id, "Obstruction detected!") // For debugging
				} else {
					logEvent(id, "Obstruction cleared") // For debugging
				}
				localStateCh <- state
			}

		// 6) Motor reinitialized
		case motor := <-motorCh:
			if state.Motorstatus {
				logEvent(id, "Motor power restored") // For debugging
				// fmt.Println("Regained motor power")
				state.Motorstatus = motor
				localStateCh <- state
			}
		}
	}
}
