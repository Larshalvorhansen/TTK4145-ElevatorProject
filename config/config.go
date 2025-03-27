package config

import (
	"time"
)

const (
	NumFloors    = 4
	NumElevators = 3
	NumButtons   = 3

	MessagePort = 20017

	BufferSize       = 1024
	ElevatorChBuffer = 16

	DisconnectTime           = 1 * time.Second
	DoorOpenDuration         = 3 * time.Second
	WatchdogTime             = 3500 * time.Millisecond
	PeerBcastInterval        = 15 * time.Millisecond
	SharedStateBcastInterval = 15 * time.Millisecond
	HardwarePollRate         = 20 * time.Millisecond
)
