package config

import (
	"time"
)

const(
	NumFloors = 4
	NumElevators = 3
	
	DisconnectTime = 2 * time.Second
	UpdateIntervalTime = 40 * time.Millisecond
)