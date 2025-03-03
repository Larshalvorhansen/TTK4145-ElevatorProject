package elevator

import (
	"Driver-go/config"
	"Driver-go/elevio"
	"time"
)

type DoorState int

const (
	Closed      DoorState = iota // Door is fully closed
	InCountDown                  // Door is open and counting down
	Obstructed                   // Door is obstructed
)

// Door controls the elevator door operation
func Door(
	doorClosedC chan<- bool, // SendoorState signal when the door is closed
	doorOpenC <-chan bool, // Receives signal to open the door
	obstructedC chan<- bool, // SendoorState signal when an obstruction is detected
) {
	elevio.SetDoorOpenLamp(false) // Ensure door lamp is off initially
	obstructionC := make(chan bool)
	go elevio.PollObstructionSwitch(obstructionC) // Continuously check for obstructions

	obstruction := false
	doorState := Closed
	timeCounter := time.NewTimer(0)
	<-timeCounter.C    // Drain initial timer event
	timeCounter.Stop() // Stop unused timer

	for {
		select {
		case obstruction = <-obstructionC: // Handle obstruction state
			if !obstruction && doorState == Obstructed {
				elevio.SetDoorOpenLamp(false) // Close door when obstruction is cleared
				doorClosedC <- true
				doorState = Closed
			}
			obstructedC <- obstruction // Notify about obstruction status

		case <-doorOpenC: // Handle door opening request
			if obstruction {
				obstructedC <- true // Notify about obstruction
			}
			if doorState == Closed || doorState == Obstructed {
				elevio.SetDoorOpenLamp(true) // Open the door
			}
			// Reset countdown timer safely
			if !timeCounter.Stop() {
				<-timeCounter.C
			}
			timeCounter.Reset(config.DoorOpenDuration)
			doorState = InCountDown

		case <-timeCounter.C: // Handle countdown expiration
			if doorState != InCountDown {
				panic("Door state not implemented") // Prevent undefined states
			}
			if obstruction {
				doorState = Obstructed // Stay open if obstructed
			} else {
				elevio.SetDoorOpenLamp(false) // Close the door
				doorClosedC <- true
				doorState = Closed
			}
		}
	}
}
