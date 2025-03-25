package config

import (
	"time"
)

const (
	// ---------------- Elevator setup ----------------
	NumFloors    = 4
	NumElevators = 3
	NumButtons   = 3

	// ------------------ Networking ------------------
	MessagePort = 20017

	// ----------------- Buffer sizes -----------------
	BufferSize       = 1024
	ElevatorChBuffer = 16

	// -------------------- Timing --------------------
	DisconnectTime    = 1 * time.Second
	DoorOpenDuration  = 3 * time.Second
	WatchdogTime      = 4 * time.Second
	PeerBcastInterval = 15 * time.Millisecond
	CoordinatorTick   = 15 * time.Millisecond
	HardwarePollRate  = 20 * time.Millisecond
)
