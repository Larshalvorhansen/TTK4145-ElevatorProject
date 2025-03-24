package elevator

import (
	"Driver-go/config"
	"Driver-go/hardware"
	"time"
)

type doorState int

const (
	closed doorState = iota
	inCountDown
	obstructed
)

func DoorLogic(
	doorClosedCh chan<- bool,
	doorOpenCh <-chan bool,
	obstructedCh chan<- bool,
) {
	hardware.SetDoorOpenLamp(false)

	obstructionCh := make(chan bool)
	go hardware.PollObstructionSwitch(obstructionCh)

	// Init state
	obstruction := false
	doorState := closed

	timeCounter := time.NewTimer(time.Hour)
	timeCounter.Stop()

	for {
		select {
		case obstruction = <-obstructionCh:
			if !obstruction && doorState == obstructed {
				hardware.SetDoorOpenLamp(false)
				doorClosedCh <- true
				doorState = closed
			}
			if obstruction {
				obstructedCh <- true
			} else {
				obstructedCh <- false
			}

		case <-doorOpenCh:
			if obstruction {
				obstructedCh <- true
			}
			switch doorState {
			case closed:
				hardware.SetDoorOpenLamp(true)
				timeCounter = time.NewTimer(config.DoorOpenDuration)
				doorState = inCountDown
			case inCountDown:
				timeCounter = time.NewTimer(config.DoorOpenDuration)

			case obstructed:
				timeCounter = time.NewTimer(config.DoorOpenDuration)
				doorState = inCountDown
			default:
				panic("DoorLogic state not implemented")
			}
		case <-timeCounter.C:
			if doorState != inCountDown {
				panic("DoorLogic state not implemented")
			}
			if obstruction {
				doorState = obstructed
			} else {
				hardware.SetDoorOpenLamp(false)
				doorClosedCh <- true
				doorState = closed
			}
		}
	}
}
