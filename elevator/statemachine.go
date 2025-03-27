package elevator

import (
	"elevator-project/config"
	"elevator-project/hardware"
	"fmt"
	"time"
)

func Elevator(
	localID int,
	localStateCh chan<- State,
	newOrderCh <-chan Orders,
	orderDeliveredCh chan<- hardware.ButtonEvent,
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
		case orders = <-newOrderCh:
			switch state.Behaviour {
			case Idle:
				switch {
				case orders[state.Floor][state.Direction] || orders[state.Floor][hardware.BT_Cab]:
					state.OpenDoorAndDeliverOrders(doorOpenCh, orders, orderDeliveredCh)
					localStateCh <- state

				case orders[state.Floor][state.Direction.FlipDirection()]:
					state.Direction = state.Direction.FlipDirection()
					state.OpenDoorAndDeliverOrders(doorOpenCh, orders, orderDeliveredCh)
					localStateCh <- state

				case orders.OrderInDirection(state.Floor, state.Direction):
					state.StartMoving(localStateCh, &motorTimer, motorActiveCh)

				case orders.OrderInDirection(state.Floor, state.Direction.FlipDirection()):
					state.Direction = state.Direction.FlipDirection()
					state.StartMoving(localStateCh, &motorTimer, motorActiveCh)
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
				panic("Unexpected state in newOrderCh case")
			}

		case state.Floor = <-floorEnteredCh:
			fmt.Printf("[Elevator %d] Detected floor %d\n", localID, state.Floor)
			hardware.SetFloorIndicator(state.Floor)
			motorTimer.Stop()
			motorActiveCh <- false

			switch state.Behaviour {
			case Moving:
				switch {
				case orders[state.Floor][state.Direction]:
					state.StopAndOpenDoor(doorOpenCh, orders, orderDeliveredCh)

				case orders[state.Floor][hardware.BT_Cab] && orders.OrderInDirection(state.Floor, state.Direction):
					state.StopAndOpenDoor(doorOpenCh, orders, orderDeliveredCh)

				case orders[state.Floor][hardware.BT_Cab] && !orders[state.Floor][state.Direction.FlipDirection()]:
					state.StopAndOpenDoor(doorOpenCh, orders, orderDeliveredCh)

				case orders.OrderInDirection(state.Floor, state.Direction):
					motorTimer = time.NewTimer(config.WatchdogTime)
					motorActiveCh <- false

				case orders[state.Floor][state.Direction.FlipDirection()]:
					state.Direction = state.Direction.FlipDirection()
					state.StopAndOpenDoor(doorOpenCh, orders, orderDeliveredCh)

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
				panic("Floor event received while in invalid state")
			}
			localStateCh <- state

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

		case <-doorClosedCh:
			fmt.Printf("[Elevator %d] Door closed at floor %d\n", localID, state.Floor)
			switch state.Behaviour {
			case DoorOpen:
				switch {
				case orders.OrderInDirection(state.Floor, state.Direction):
					state.StartMoving(localStateCh, &motorTimer, motorActiveCh)

				case orders[state.Floor][state.Direction.FlipDirection()]:
					doorOpenCh <- true
					state.Direction = state.Direction.FlipDirection()
					SendCompletedOrders(state.Floor, state.Direction, orders, orderDeliveredCh)
					localStateCh <- state

				case orders.OrderInDirection(state.Floor, state.Direction.FlipDirection()):
					state.Direction = state.Direction.FlipDirection()
					state.StartMoving(localStateCh, &motorTimer, motorActiveCh)

				default:
					state.Behaviour = Idle
					localStateCh <- state
				}

			default:
				panic("Unexpected behaviour state on door closed event")
			}

		case <-motorTimer.C:
			if !state.MotorPowerLost {
				fmt.Printf("[Elevator %d] WARNING: Lost motor power!\n", localID)
				state.MotorPowerLost = true
				localStateCh <- state
			}

		case motor := <-motorActiveCh:
			if state.MotorPowerLost {
				fmt.Printf("[Elevator %d] Motor power restored\n", localID)
				state.MotorPowerLost = motor
				localStateCh <- state
			}
		}
	}
}
