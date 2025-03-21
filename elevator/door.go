package elevator

import (
	"Driver-go/config"
	"Driver-go/hardware"
	"time"
)

type DoorState int

const (
	Closed DoorState = iota
	InCountDown
	Obstructed
)

func Door(
	doorClosedC chan<- bool,
	doorOpenC <-chan bool,
	obstructedC chan<- bool,
) {
	hardware.SetDoorOpenLamp(false)

	obstructionC := make(chan bool)
	go hardware.PollObstructionSwitch(obstructionC)

	// Init state
	obstruction := false
	doorState := Closed

	timeCounter := time.NewTimer(time.Hour)
	timeCounter.Stop()

	for {
		select {
		case obstruction = <-obstructionC:
			if !obstruction && doorState == Obstructed {
				hardware.SetDoorOpenLamp(false)
				doorClosedC <- true
				doorState = Closed
			}
			if obstruction {
				obstructedC <- true
			} else {
				obstructedC <- false
			}

		case <-doorOpenC:
			if obstruction {
				obstructedC <- true
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
				panic("Door state not implemented")
			}
		case <-timeCounter.C:
			if doorState != InCountDown {
				panic("Door state not implemented")
			}
			if obstruction {
				doorState = Obstructed
			} else {
				hardware.SetDoorOpenLamp(false)
				doorClosedC <- true
				doorState = Closed
			}
		}
	}
}
