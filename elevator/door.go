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
	obstruction := false
	doorState := Closed
	hardware.SetDoorOpenLamp(false)

	obstructionCh := make(chan bool)
	go hardware.PollObstructionSwitch(obstructionCh)

	doorOpenTimer := time.NewTimer(time.Hour)
	doorOpenTimer.Stop()

	for {
		select {

		// ------------------------------------- Door is open -------------------------------------
		case <-doorOpenCh:
			if obstruction {
				obstructedCh <- true
			}
			switch doorState {
			case Closed:
				hardware.SetDoorOpenLamp(true)
				doorOpenTimer = time.NewTimer(config.DoorOpenDuration)
				doorState = InCountDown
			case InCountDown:
				doorOpenTimer = time.NewTimer(config.DoorOpenDuration)

			case Obstructed:
				doorOpenTimer = time.NewTimer(config.DoorOpenDuration)
				doorState = InCountDown
			default:
				panic("Unexpected elevator state in DoorLogic: state transition not implemented")
			}

		// ------------------------------ Obstruction switch pressed ------------------------------
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

		// -------------------------------- Door open timer expires -------------------------------
		case <-doorOpenTimer.C:
			if doorState != InCountDown {
				panic("Unexpected elevator state in DoorLogic: state transition not implemented")
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
