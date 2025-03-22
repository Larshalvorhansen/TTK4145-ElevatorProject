package elevator

import (
	"Driver-go/config"
	"Driver-go/hardware"
	"fmt"
	"time"
)

// logEvent: En hjelpefunksjon for å skrive ut viktige hendelser
func logEvent(format string, args ...interface{}) {
	fmt.Printf("[Elevator] "+format+"\n", args...)
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
	newOrderC <-chan Orders,
	deliveredOrderC chan<- hardware.ButtonEvent,
	newStateC chan<- State,
) {
	doorOpenC := make(chan bool, config.ElevatorChBuffer)
	doorClosedC := make(chan bool, config.ElevatorChBuffer)
	floorEnteredC := make(chan int)
	obstructedC := make(chan bool, config.ElevatorChBuffer)
	motorC := make(chan bool, config.ElevatorChBuffer)

	go DoorLogic(doorClosedC, doorOpenC, obstructedC)
	go hardware.PollFloorSensor(floorEnteredC)

	hardware.SetMotorDirection(hardware.MD_Down)
	state := State{Direction: Down, Behaviour: Moving}

	var orders Orders

	motorTimer := time.NewTimer(config.WatchdogTime)
	motorTimer.Stop()

	logEvent("Starting elevator in state: %s", state.Behaviour.ToString()) // For debugging

	for {
		select {

		// 1) New order incoming
		case orders = <-newOrderC:
			logEvent("New order received") // For debugging
			switch state.Behaviour {
			case Idle:
				switch {
				case orders[state.Floor][state.Direction] || orders[state.Floor][hardware.BT_Cab]:
					doorOpenC <- true
					SendOrderDone(state.Floor, state.Direction, orders, deliveredOrderC)
					state.Behaviour = DoorOpen
					newStateC <- state

				case orders[state.Floor][state.Direction.FlipDirection()]:
					doorOpenC <- true
					state.Direction = state.Direction.FlipDirection()
					SendOrderDone(state.Floor, state.Direction, orders, deliveredOrderC)
					state.Behaviour = DoorOpen
					newStateC <- state

				case orders.OrderInDirection(state.Floor, state.Direction):
					hardware.SetMotorDirection(state.Direction.ToMotorDirection())
					state.Behaviour = Moving
					newStateC <- state
					motorTimer = time.NewTimer(config.WatchdogTime)
					motorC <- false

				case orders.OrderInDirection(state.Floor, state.Direction.FlipDirection()):
					state.Direction = state.Direction.FlipDirection()
					hardware.SetMotorDirection(state.Direction.ToMotorDirection())
					state.Behaviour = Moving
					newStateC <- state
					motorTimer = time.NewTimer(config.WatchdogTime)
					motorC <- false
				default:
				}

			case DoorOpen:
				switch {
				case orders[state.Floor][hardware.BT_Cab] || orders[state.Floor][state.Direction]:
					doorOpenC <- true
					SendOrderDone(state.Floor, state.Direction, orders, deliveredOrderC)
				}

			case Moving:

			default:
				panic("Orders in wrong state")
			}

		// 2) DoorLogic closes
		case <-doorClosedC:
			logEvent("DoorLogic closed at floor %d", state.Floor) // For debugging
			switch state.Behaviour {
			case DoorOpen:
				switch {
				case orders.OrderInDirection(state.Floor, state.Direction):
					hardware.SetMotorDirection(state.Direction.ToMotorDirection())
					state.Behaviour = Moving
					motorTimer = time.NewTimer(config.WatchdogTime)
					motorC <- false
					newStateC <- state

				case orders[state.Floor][state.Direction.FlipDirection()]:
					doorOpenC <- true
					state.Direction = state.Direction.FlipDirection()
					SendOrderDone(state.Floor, state.Direction, orders, deliveredOrderC)
					newStateC <- state

				case orders.OrderInDirection(state.Floor, state.Direction.FlipDirection()):
					state.Direction = state.Direction.FlipDirection()
					hardware.SetMotorDirection(state.Direction.ToMotorDirection())
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

		// 3) Elevator finds new floor
		case state.Floor = <-floorEnteredC:
			logEvent("Arrived at floor %d", state.Floor) // For debugging
			hardware.SetFloorIndicator(state.Floor)
			motorTimer.Stop()
			motorC <- false

			switch state.Behaviour {
			case Moving:
				switch {
				case orders[state.Floor][state.Direction]:
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
					motorTimer = time.NewTimer(config.WatchdogTime)
					motorC <- false

				case orders[state.Floor][state.Direction.FlipDirection()]:
					hardware.SetMotorDirection(hardware.MD_Stop)
					doorOpenC <- true
					state.Direction = state.Direction.FlipDirection()
					SendOrderDone(state.Floor, state.Direction, orders, deliveredOrderC)
					state.Behaviour = DoorOpen

				case orders.OrderInDirection(state.Floor, state.Direction.FlipDirection()):
					state.Direction = state.Direction.FlipDirection()
					hardware.SetMotorDirection(state.Direction.ToMotorDirection())
					motorTimer = time.NewTimer(config.WatchdogTime)
					motorC <- false

				default:
					hardware.SetMotorDirection(hardware.MD_Stop)
					state.Behaviour = Idle
				}

			default:
				panic("FloorEntered in wrong state")
			}
			newStateC <- state

		// 4) MOTOR‐WATCHDOG time gone out
		case <-motorTimer.C:
			if !state.Motorstatus {
				logEvent("WARNING: Lost motor power!") // For debugging
				// fmt.Println("Lost motor power")
				state.Motorstatus = true
				newStateC <- state
			}

		// 5) Obstruction
		case obstruction := <-obstructedC:
			if obstruction != state.Obstructed {
				state.Obstructed = obstruction
				if obstruction {
					logEvent("Obstruction detected!") // For debugging
				} else {
					logEvent("Obstruction cleared") // For debugging
				}
				newStateC <- state
			}

		// 6) Motor reinitialized
		case motor := <-motorC:
			if state.Motorstatus {
				logEvent("Motor power restored") // For debugging
				// fmt.Println("Regained motor power")
				state.Motorstatus = motor
				newStateC <- state
			}
		}
	}
}
