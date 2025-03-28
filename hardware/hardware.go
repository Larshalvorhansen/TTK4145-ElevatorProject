// The following implementation is based on the driver provided in TTK4145's project resources:
// https://github.com/TTK4145/driver-go/blob/master/elevio/elevator_io.go
// Modifications were made to adapt it to this project's architecture and requirements.

//TODO: Chech if the commented out functions are needed or not. If not, remove them.

package hardware

import (
	"elevator-project/config"
	"fmt"
	"net"
	"sync"
	"time"
)

var initialized bool = false
var mtx sync.Mutex
var conn net.Conn

type MotorDirection int

const (
	MD_Up   MotorDirection = 1
	MD_Down                = -1
	MD_Stop                = 0
)

type ButtonType int

const (
	BT_HallUp   ButtonType = 0
	BT_HallDown            = 1
	BT_Cab                 = 2
)

type ButtonEvent struct {
	Floor  int
	Button ButtonType
}

func Init(addr string) {
	if initialized {
		fmt.Println("Driver already initialized!")
		return
	}
	mtx = sync.Mutex{}
	var err error
	conn, err = net.Dial("tcp", addr)
	if err != nil {
		panic(err.Error())
	}
	initialized = true
}

func SetMotorDirection(dir MotorDirection) {
	write([4]byte{1, byte(dir), 0, 0})
}

func SetButtonLamp(button ButtonType, floor int, value bool) {
	write([4]byte{2, byte(button), byte(floor), toByte(value)})
}

func SetFloorIndicator(floor int) {
	write([4]byte{3, byte(floor), 0, 0})
}

func SetDoorOpenLamp(value bool) {
	write([4]byte{4, toByte(value), 0, 0})
}

// func SetStopLamp(value bool) {
// 	write([4]byte{5, toByte(value), 0, 0})
// }

func PollButtons(receiver chan<- ButtonEvent) {
	prev := make([][3]bool, config.NumFloors)
	for {
		time.Sleep(config.HardwarePollRate)
		for f := 0; f < config.NumFloors; f++ {
			for b := ButtonType(0); b < 3; b++ {
				v := GetButton(b, f)
				if v != prev[f][b] && v != false {
					receiver <- ButtonEvent{f, ButtonType(b)}
				}
				prev[f][b] = v
			}
		}
	}
}

func PollFloorSensor(receiver chan<- int) {
	prev := -1
	for {
		time.Sleep(config.HardwarePollRate)
		v := GetFloor()
		if v != prev && v != -1 {
			receiver <- v
		}
		prev = v
	}
}

// func PollStopButton(receiver chan<- bool) {
// 	prev := false
// 	for {
// 		time.Sleep(config.HardwarePollRate)
// 		v := GetStop()
// 		if v != prev {
// 			receiver <- v
// 		}
// 		prev = v
// 	}
// }

func PollObstructionSwitch(receiver chan<- bool) {
	prev := false
	for {
		time.Sleep(config.HardwarePollRate)
		v := GetObstruction()
		if v != prev {
			receiver <- v
		}
		prev = v
	}
}

func GetButton(button ButtonType, floor int) bool {
	a := read([4]byte{6, byte(button), byte(floor), 0})
	return toBool(a[1])
}

func GetFloor() int {
	a := read([4]byte{7, 0, 0, 0})
	if a[1] != 0 {
		return int(a[2])
	} else {
		return -1
	}
}

// func GetStop() bool {
// 	a := read([4]byte{8, 0, 0, 0})
// 	return toBool(a[1])
// }

func GetObstruction() bool {
	a := read([4]byte{9, 0, 0, 0})
	return toBool(a[1])
}

func read(in [4]byte) [4]byte {
	mtx.Lock()
	defer mtx.Unlock()

	_, err := conn.Write(in[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}

	var out [4]byte
	_, err = conn.Read(out[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}

	return out
}

func write(in [4]byte) {
	mtx.Lock()
	defer mtx.Unlock()

	_, err := conn.Write(in[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}
}

func toByte(a bool) byte {
	var b byte = 0
	if a {
		b = 1
	}
	return b
}

func toBool(a byte) bool {
	var b bool = false
	if a != 0 {
		b = true
	}
	return b
}
