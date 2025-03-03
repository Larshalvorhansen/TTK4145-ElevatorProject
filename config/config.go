package config

import (
	"time"
)

const (
	NumFloors    = 4
	NumElevators = 3
	NumButtons   = 3
	Buffer       = 1024

	DisconnectTime     = 2 * time.Second
	DoorOpenDuration   = 3 * time.Second
	UpdateIntervalTime = 40 * time.Millisecond
	WatchdogTime       = 4 * time.Second
)
