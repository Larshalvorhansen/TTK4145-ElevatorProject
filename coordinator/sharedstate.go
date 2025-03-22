package coordinator

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

type SharedState struct {
	SeqNum       int
	Origin       int
	Ackmap       [config.NumElevators]AckStatus
	HallRequests [config.NumFloors][2]bool
	States       [config.NumElevators]LocalState
}

func (ss *SharedState) addOrder(newOrder hardware.ButtonEvent, id int) {
	if newOrder.Button == hardware.BT_Cab {
		ss.States[id].CabRequests[newOrder.Floor] = true
	} else {
		ss.HallRequests[newOrder.Floor][newOrder.Button] = true
	}
}

func (ss *SharedState) addCabCall(newOrder hardware.ButtonEvent, id int) {
	if newOrder.Button == hardware.BT_Cab {
		ss.States[id].CabRequests[newOrder.Floor] = true
	}
}

func (ss *SharedState) removeOrder(deliveredOrder hardware.ButtonEvent, id int) {
	if deliveredOrder.Button == hardware.BT_Cab {
		ss.States[id].CabRequests[deliveredOrder.Floor] = false
	} else {
		ss.HallRequests[deliveredOrder.Floor][deliveredOrder.Button] = false
	}
}

func (ss *SharedState) updateState(newState elevator.State, id int) {
	ss.States[id] = LocalState{
		State:       newState,
		CabRequests: ss.States[id].CabRequests,
	}
}

func (ss *SharedState) fullyAcked(id int) bool {
	if ss.Ackmap[id] == NotAvailable {
		return false
	}
	for index := range ss.Ackmap {
		if ss.Ackmap[index] == NotAcked {
			return false
		}
	}
	return true
}

func (oldCs SharedState) equals(newCs SharedState) bool {
	oldCs.Ackmap = [config.NumElevators]AckStatus{}
	newCs.Ackmap = [config.NumElevators]AckStatus{}
	return reflect.DeepEqual(oldCs, newCs)
}

func (ss *SharedState) makeLostPeersUnavailable(peers peers.PeerUpdate) {
	for _, id := range peers.Lost {
		ss.Ackmap[id] = NotAvailable
	}
}

func (ss *SharedState) makeOthersUnavailable(id int) {
	for elev := range ss.Ackmap {
		if elev != id {
			ss.Ackmap[elev] = NotAvailable
		}
	}
}

func (ss *SharedState) prepNewCs(id int) {
	ss.SeqNum++
	ss.Origin = id
	for id := range ss.Ackmap {
		if ss.Ackmap[id] == Acked {
			ss.Ackmap[id] = NotAcked
		}
	}
}
