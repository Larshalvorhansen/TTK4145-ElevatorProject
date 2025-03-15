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

type CommonState struct {
	SeqNum       int
	Origin       int
	Ackmap       [config.NumElevators]AckStatus
	HallRequests [config.NumFloors][2]bool
	States       [config.NumElevators]LocalState
}

func (cs *CommonState) addOrder(newOrder hardware.ButtonEvent, id int) {
	if newOrder.Button == hardware.BT_Cab {
		cs.States[id].CabRequests[newOrder.Floor] = true
	} else {
		cs.HallRequests[newOrder.Floor][newOrder.Button] = true
	}
}

func (cs *CommonState) addCabCall(newOrder hardware.ButtonEvent, id int) {
	if newOrder.Button == hardware.BT_Cab {
		cs.States[id].CabRequests[newOrder.Floor] = true
	}
}

func (cs *CommonState) removeOrder(deliveredOrder hardware.ButtonEvent, id int) {
	if deliveredOrder.Button == hardware.BT_Cab {
		cs.States[id].CabRequests[deliveredOrder.Floor] = false
	} else {
		cs.HallRequests[deliveredOrder.Floor][deliveredOrder.Button] = false
	}
}

func (cs *CommonState) updateState(newState elevator.State, id int) {
	cs.States[id] = LocalState{
		State:       newState,
		CabRequests: cs.States[id].CabRequests,
	}
}

func (cs *CommonState) fullyAcked(id int) bool {
	if cs.Ackmap[id] == NotAvailable {
		return false
	}
	for index := range cs.Ackmap {
		if cs.Ackmap[index] == NotAcked {
			return false
		}
	}
	return true
}

func (oldCs CommonState) equals(newCs CommonState) bool {
	oldCs.Ackmap = [config.NumElevators]AckStatus{}
	newCs.Ackmap = [config.NumElevators]AckStatus{}
	return reflect.DeepEqual(oldCs, newCs)
}

func (cs *CommonState) makeLostPeersUnavailable(peers peers.PeerUpdate) {
	for _, id := range peers.Lost {
		cs.Ackmap[id] = NotAvailable
	}
}

func (cs *CommonState) makeOthersUnavailable(id int) {
	for elev := range cs.Ackmap {
		if elev != id {
			cs.Ackmap[elev] = NotAvailable
		}
	}
}

func (cs *CommonState) prepNewCs(id int) {
	cs.SeqNum++
	cs.Origin = id
	for id := range cs.Ackmap {
		if cs.Ackmap[id] == Acked {
			cs.Ackmap[id] = NotAcked
		}
	}
}
