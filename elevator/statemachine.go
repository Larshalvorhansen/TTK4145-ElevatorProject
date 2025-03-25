// TODO: Check if this is ok with group, if not copy state.go and paste under "import()" to get the old version back

package elevator

import (
	"elevator-project/config"
	"elevator-project/hardware"
	"fmt"
	"time"
)

func Elevator(
	newOrderCh <-chan Orders,
	orderDeliveredCh chan<- hardware.ButtonEvent,
	localStateCh chan<- State,
	localID int,
) {
	doorOpenCh := make(chan bool, config.ElevatorChBuffer)
	doorClosedCh := make(chan bool, config.ElevatorChBuffer)
	floorEnteredCh := make(chan int)
	obstructedCh := make(chan bool, config.ElevatorChBuffer)
	motorActiveCh := make(chan bool, config.ElevatorChBuffer)

	go DoorLogic(doorClosedCh, doorOpenCh, obstructedCh)
	go hardware.PollFloorSensor(floorEnteredCh)

	hardware.SetMotorDirection(hardware.MD_Down)
	state := State{Direction: Down, Behaviour: Moving}

	var orders Orders

	motorTimer := time.NewTimer(config.WatchdogTime)
	motorTimer.Stop()

	for {
		select {

		// -------------------------------------- New order incoming --------------------------------------
		case orders = <-newOrderCh:
			switch state.Behaviour {
			case Idle:
				switch {
				case orders[state.Floor][state.Direction] || orders[state.Floor][hardware.BT_Cab]:
					doorOpenCh <- true
					SendCompletedOrders(state.Floor, state.Direction, orders, orderDeliveredCh)
					state.Behaviour = DoorOpen
					localStateCh <- state

				case orders[state.Floor][state.Direction.FlipDirection()]:
					doorOpenCh <- true
					state.Direction = state.Direction.FlipDirection()
					SendCompletedOrders(state.Floor, state.Direction, orders, orderDeliveredCh)
					state.Behaviour = DoorOpen
					localStateCh <- state

				case orders.OrderInDirection(state.Floor, state.Direction):
					hardware.SetMotorDirection(state.Direction.ToMotorDirection())
					state.Behaviour = Moving
					localStateCh <- state
					motorTimer = time.NewTimer(config.WatchdogTime)
					motorActiveCh <- false

				case orders.OrderInDirection(state.Floor, state.Direction.FlipDirection()):
					state.Direction = state.Direction.FlipDirection()
					hardware.SetMotorDirection(state.Direction.ToMotorDirection())
					state.Behaviour = Moving
					localStateCh <- state
					motorTimer = time.NewTimer(config.WatchdogTime)
					motorActiveCh <- false
				default:
				}

			case DoorOpen:
				switch {
				case orders[state.Floor][hardware.BT_Cab] || orders[state.Floor][state.Direction]:
					doorOpenCh <- true
					SendCompletedOrders(state.Floor, state.Direction, orders, orderDeliveredCh)
				}

			case Moving:

			default:
				panic("Orders in wrong state")
			}

		// ----------------------------------- Elevator finds new floor -----------------------------------
		case state.Floor = <-floorEnteredCh:
			fmt.Printf("[Elevator %d] Detected floor %d\n", localID, state.Floor)
			hardware.SetFloorIndicator(state.Floor)
			motorTimer.Stop()
			motorActiveCh <- false

			switch state.Behaviour {
			case Moving:
				switch {
				case orders[state.Floor][state.Direction]:
					hardware.SetMotorDirection(hardware.MD_Stop)
					doorOpenCh <- true
					SendCompletedOrders(state.Floor, state.Direction, orders, orderDeliveredCh)
					state.Behaviour = DoorOpen

				case orders[state.Floor][hardware.BT_Cab] && orders.OrderInDirection(state.Floor, state.Direction):
					hardware.SetMotorDirection(hardware.MD_Stop)
					doorOpenCh <- true
					SendCompletedOrders(state.Floor, state.Direction, orders, orderDeliveredCh)
					state.Behaviour = DoorOpen

				case orders[state.Floor][hardware.BT_Cab] && !orders[state.Floor][state.Direction.FlipDirection()]:
					hardware.SetMotorDirection(hardware.MD_Stop)
					doorOpenCh <- true
					SendCompletedOrders(state.Floor, state.Direction, orders, orderDeliveredCh)
					state.Behaviour = DoorOpen

				case orders.OrderInDirection(state.Floor, state.Direction):
					motorTimer = time.NewTimer(config.WatchdogTime)
					motorActiveCh <- false

				case orders[state.Floor][state.Direction.FlipDirection()]:
					hardware.SetMotorDirection(hardware.MD_Stop)
					doorOpenCh <- true
					state.Direction = state.Direction.FlipDirection()
					SendCompletedOrders(state.Floor, state.Direction, orders, orderDeliveredCh)
					state.Behaviour = DoorOpen

				case orders.OrderInDirection(state.Floor, state.Direction.FlipDirection()):
					state.Direction = state.Direction.FlipDirection()
					hardware.SetMotorDirection(state.Direction.ToMotorDirection())
					motorTimer = time.NewTimer(config.WatchdogTime)
					motorActiveCh <- false

				default:
					hardware.SetMotorDirection(hardware.MD_Stop)
					state.Behaviour = Idle
				}

			default:
				panic("FloorEntered in wrong state")
			}
			localStateCh <- state

		// ------------------------------------------ Obstruction -----------------------------------------
		case obstruction := <-obstructedCh:
			if obstruction != state.Obstructed {
				state.Obstructed = obstruction
				if obstruction {
					fmt.Printf("[Elevator %d] Obstruction detected!\n", localID)

				} else {
					fmt.Printf("[Elevator %d] Obstruction cleared!\n", localID)
				}
				localStateCh <- state
			}

		// ----------------------------------------- Door closes ------------------------------------------
		case <-doorClosedCh:
			fmt.Printf("[Elevator %d] Door closed at floor %d\n", localID, state.Floor)
			switch state.Behaviour {
			case DoorOpen:
				switch {
				case orders.OrderInDirection(state.Floor, state.Direction):
					hardware.SetMotorDirection(state.Direction.ToMotorDirection())
					state.Behaviour = Moving
					motorTimer = time.NewTimer(config.WatchdogTime)
					motorActiveCh <- false
					localStateCh <- state

				case orders[state.Floor][state.Direction.FlipDirection()]:
					doorOpenCh <- true
					state.Direction = state.Direction.FlipDirection()
					SendCompletedOrders(state.Floor, state.Direction, orders, orderDeliveredCh)
					localStateCh <- state

				case orders.OrderInDirection(state.Floor, state.Direction.FlipDirection()):
					state.Direction = state.Direction.FlipDirection()
					hardware.SetMotorDirection(state.Direction.ToMotorDirection())
					state.Behaviour = Moving
					motorTimer = time.NewTimer(config.WatchdogTime)
					motorActiveCh <- false
					localStateCh <- state

				default:
					state.Behaviour = Idle
					localStateCh <- state
				}

			default:
				panic("Door in wrong state")
			}

		// --------------------------------- MOTOR‐WATCHDOG time gone out ---------------------------------
		case <-motorTimer.C:
			if !state.Motorstatus {
				fmt.Printf("[Elevator %d] WARNING: Lost motor power!\n", localID)
				state.Motorstatus = true
				localStateCh <- state
			}

		// -------------------------------------- Motor reinitialized -------------------------------------
		case motor := <-motorActiveCh:
			if state.Motorstatus {
				fmt.Printf("[Elevator %d] Motor power restored\n", localID)
				state.Motorstatus = motor
				localStateCh <- state
			}
		}
	}
}
