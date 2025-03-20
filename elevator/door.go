// TODO: check iota constants, if they are exported capitalize firt letter, if not, keep it lowercase in first letter

package elevator

import (
	"Driver-go/config"
	"Driver-go/hardware"
	"log"
	"time"
)

type DoorState int

const (
	Closed DoorState = iota
	InCountDown
	Obstructed
)

func DoorLogic(
	doorClosedC chan<- bool,
	doorOpenC <-chan bool,
	obstructedC chan<- bool,
) {
	hardware.SetDoorOpenLamp(false)

	obstructionC := make(chan bool)
	go hardware.PollObstructionSwitch(obstructionC)

	obstruction := false
	doorState := Closed

	timer := time.NewTimer(time.Hour)
	timer.Stop()

	for {
		select {
		case obstruction = <-obstructionC:
			if !obstruction && doorState == Obstructed {
				log.Println("Door no longer obstructed, closing now")
				hardware.SetDoorOpenLamp(false)
				doorClosedC <- true
				doorState = Closed
			}
			obstructedC <- obstruction

		case <-doorOpenC:
			if obstruction {
				obstructedC <- true
			}
			switch doorState {
			case Closed:
				hardware.SetDoorOpenLamp(true)
				timer = time.NewTimer(config.DoorOpenDuration)
				doorState = InCountDown

			case InCountDown:
				timer.Stop()
				timer = time.NewTimer(config.DoorOpenDuration)

			case Obstructed:
				log.Println("Door was obstructed, resuming countdown")
				timer.Stop()
				timer = time.NewTimer(config.DoorOpenDuration)
				doorState = InCountDown

			default:
				panic("Door state not implemented")
			}

		case <-timer.C:
			if doorState != InCountDown {
				panic("Door state not implemented")
			}
			if obstruction {
				log.Println("Door is obstructed, keeping it open")
				doorState = Obstructed
			} else {
				hardware.SetDoorOpenLamp(false)
				doorClosedC <- true
				doorState = Closed
			}
		}
	}
}
