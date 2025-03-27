// TODO: Change layout and logical order of the code. Also change panics prints

package elevator

import (
	"elevator-project/config"
	"elevator-project/hardware"
	"time"
)

type DoorState int

const (
	Closed DoorState = iota
	InCountDown
	Obstructed
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
	doorState := Closed

	timeCounter := time.NewTimer(time.Hour)
	timeCounter.Stop()

	for {
		select {
		case obstruction = <-obstructionCh:
			if !obstruction && doorState == Obstructed {
				hardware.SetDoorOpenLamp(false)
				doorClosedCh <- true
				doorState = Closed
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
			case Closed:
				hardware.SetDoorOpenLamp(true)
				timeCounter = time.NewTimer(config.DoorOpenDuration)
				doorState = InCountDown
			case InCountDown:
				timeCounter = time.NewTimer(config.DoorOpenDuration)

			case Obstructed:
				timeCounter = time.NewTimer(config.DoorOpenDuration)
				doorState = InCountDown
			default:
				panic("DoorLogic state not implemented")
			}
		case <-timeCounter.C:
			if doorState != InCountDown {
				panic("DoorLogic state not implemented")
			}
			if obstruction {
				doorState = Obstructed
			} else {
				hardware.SetDoorOpenLamp(false)
				doorClosedCh <- true
				doorState = Closed
			}
		}
	}
}
