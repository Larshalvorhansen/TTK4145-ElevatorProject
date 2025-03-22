package config

import (
	"time"
)

const (
	// Change these parameters for custom configurations
	NumFloors       = 4
	NumElevators    = 1
	NumButtons      = 3
	PeersPortNumber = 20017
	BcastPortNumber = 20017

	BufferSize = 1024

	DisconnectTime   = 1 * time.Second
	DoorOpenDuration = 3 * time.Second
	WatchdogTime     = 4 * time.Second
	Interval         = 15 * time.Millisecond
)
