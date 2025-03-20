package config

import "time"

const (
	// Elevator setup
	NumFloors    = 4
	NumElevators = 1
	NumButtons   = 3

	// Networking
	PeersPortNumber = 58735
	BcastPortNumber = 58750

	// Channel and buffer sizes
	BufferSize         = 1024
	ElevatorChannelBuf = 16
	HardwarePollRate   = 20 * time.Millisecond

	// Timing
	DisconnectTime   = 1 * time.Second
	DoorOpenDuration = 3 * time.Second
	WatchdogTime     = 4 * time.Second
	PeerInterval     = 15 * time.Millisecond
	DistributorTick  = 15 * time.Millisecond
)
