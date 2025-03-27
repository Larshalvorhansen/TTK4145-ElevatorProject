package elevator

import (
	"elevator-project/config"
	"elevator-project/hardware"
	"time"
)

type State struct {
	Obstructed     bool
	MotorPowerLost bool
	Behaviour      Behaviour
	Floor          int
	Direction      Direction
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

func (s *State) OpenDoorAndDeliverOrders(
	doorOpenCh chan<- bool,
	orders Orders,
	orderDeliveredCh chan<- hardware.ButtonEvent,
) {
	doorOpenCh <- true
	SendCompletedOrders(s.Floor, s.Direction, orders, orderDeliveredCh)
	s.Behaviour = DoorOpen
}

func (s *State) StartMoving(
	localStateCh chan<- State,
	motorTimer **time.Timer,
	motorActiveCh chan<- bool,
) {
	hardware.SetMotorDirection(s.Direction.ToMotorDirection())
	s.Behaviour = Moving
	localStateCh <- *s
	*motorTimer = time.NewTimer(config.WatchdogTime)
	motorActiveCh <- false
}

func (s *State) StopAndOpenDoor(
	doorOpenCh chan<- bool,
	orders Orders,
	orderDeliveredCh chan<- hardware.ButtonEvent,
) {
	hardware.SetMotorDirection(hardware.MD_Stop)
	s.OpenDoorAndDeliverOrders(doorOpenCh, orders, orderDeliveredCh)
}
