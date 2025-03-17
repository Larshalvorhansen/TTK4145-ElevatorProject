package distributor

import (
	"Driver-go/config"
	"Driver-go/elevator"
	"Driver-go/hardware"
	"Driver-go/network/peers"
	"reflect"
)

type AckStatus int

const (
	NotAcked AckStatus = iota
	Acked
	NotAvailable
)

type LocalState struct {
	State       elevator.State
	CabRequests [config.NumFloors]bool
}

type SystemView struct {
	SeqNum         int
	Origin         int
	Ackmap         [config.NumElevators]AckStatus
	HallRequests   [config.NumFloors][2]bool
	ElevatorStates [config.NumElevators]LocalState
}

func (systemView *SystemView) addOrder(newOrder hardware.ButtonEvent, id int) {
	if newOrder.Button == hardware.BT_Cab {
		systemView.ElevatorStates[id].CabRequests[newOrder.Floor] = true
	} else {
		systemView.HallRequests[newOrder.Floor][newOrder.Button] = true
	}
}

func (systemView *SystemView) addCabCall(newOrder hardware.ButtonEvent, id int) {
	if newOrder.Button == hardware.BT_Cab {
		systemView.ElevatorStates[id].CabRequests[newOrder.Floor] = true
	}
}

func (systemView *SystemView) removeOrder(deliveredOrder hardware.ButtonEvent, id int) {
	if deliveredOrder.Button == hardware.BT_Cab {
		systemView.ElevatorStates[id].CabRequests[deliveredOrder.Floor] = false
	} else {
		systemView.HallRequests[deliveredOrder.Floor][deliveredOrder.Button] = false
	}
}

func (systemView *SystemView) updateState(newState elevator.State, id int) {
	systemView.ElevatorStates[id] = LocalState{
		State:       newState,
		CabRequests: systemView.ElevatorStates[id].CabRequests,
	}
}

func (systemView *SystemView) fullyAcked(id int) bool {
	if systemView.Ackmap[id] == NotAvailable {
		return false
	}
	for index := range systemView.Ackmap {
		if systemView.Ackmap[index] == NotAcked {
			return false
		}
	}
	return true
}

func (oldSystemView SystemView) equals(newSystemView SystemView) bool {
	oldSystemView.Ackmap = [config.NumElevators]AckStatus{}
	newSystemView.Ackmap = [config.NumElevators]AckStatus{}
	return reflect.DeepEqual(oldSystemView, newSystemView)
}

func (systemView *SystemView) makeLostPeersUnavailable(peers peers.PeerUpdate) {
	for _, id := range peers.Lost {
		systemView.Ackmap[id] = NotAvailable
	}
}

func (systemView *SystemView) makeOthersUnavailable(id int) {
	for elev := range systemView.Ackmap {
		if elev != id {
			systemView.Ackmap[elev] = NotAvailable
		}
	}
}

func (systemView *SystemView) prepNewSV(id int) {
	systemView.SeqNum++
	systemView.Origin = id
	for id := range systemView.Ackmap {
		if systemView.Ackmap[id] == Acked {
			systemView.Ackmap[id] = NotAcked
		}
	}
}
