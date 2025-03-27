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

	obstruction := false
	doorState := Closed

	timeCounter := time.NewTimer(time.Hour)
	timeCounter.Stop()

	for {
		select {
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
				panic("Unknown door state on doorOpen event")
			}

		case obstruction = <-obstructionCh:
			obstructedCh <- obstruction
			if !obstruction && doorState == Obstructed {
				hardware.SetDoorOpenLamp(false)
				doorClosedCh <- true
				doorState = Closed
			}

		case <-timeCounter.C:
			if doorState != InCountDown {
				panic("Timer expired in unexpected door state")
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
