package coordinator

import (
	"Driver-go/config"
	"Driver-go/elevator"
	"Driver-go/hardware"
	"Driver-go/network/peers"
	"reflect"
)

type ackStatus int

const (
	notAcked ackStatus = iota
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
	Ackmap       [config.NumElevators]ackStatus
	HallRequests [config.NumFloors][2]bool
	States       [config.NumElevators]LocalState
}

func (ss *SharedState) AddOrder(newOrder hardware.ButtonEvent, id int) {
	if newOrder.Button == hardware.BT_Cab {
		ss.States[id].CabRequests[newOrder.Floor] = true
	} else {
		ss.HallRequests[newOrder.Floor][newOrder.Button] = true
	}
}

func (ss *SharedState) AddCabCall(newOrder hardware.ButtonEvent, id int) {
	if newOrder.Button == hardware.BT_Cab {
		ss.States[id].CabRequests[newOrder.Floor] = true
	}
}

func (ss *SharedState) RemoveOrder(deliveredOrder hardware.ButtonEvent, id int) {
	if deliveredOrder.Button == hardware.BT_Cab {
		ss.States[id].CabRequests[deliveredOrder.Floor] = false
	} else {
		ss.HallRequests[deliveredOrder.Floor][deliveredOrder.Button] = false
	}
}

func (ss *SharedState) UpdateState(newState elevator.State, id int) {
	ss.States[id] = LocalState{
		State:       newState,
		CabRequests: ss.States[id].CabRequests,
	}
}

func (ss *SharedState) FullyAcked(id int) bool {
	if ss.Ackmap[id] == NotAvailable {
		return false
	}
	for index := range ss.Ackmap {
		if ss.Ackmap[index] == notAcked {
			return false
		}
	}
	return true
}

func (oldSs SharedState) Equals(newSs SharedState) bool {
	oldSs.Ackmap = [config.NumElevators]ackStatus{}
	newSs.Ackmap = [config.NumElevators]ackStatus{}
	return reflect.DeepEqual(oldSs, newSs)
}

func (ss *SharedState) MakeLostPeersUnavailable(peers peers.PeerUpdate) {
	for _, id := range peers.Lost {
		ss.Ackmap[id] = NotAvailable
	}
}

func (ss *SharedState) MakeOthersUnavailable(id int) {
	for elev := range ss.Ackmap {
		if elev != id {
			ss.Ackmap[elev] = NotAvailable
		}
	}
}

func (ss *SharedState) PrepNewSs(id int) {
	ss.SeqNum++
	ss.Origin = id
	for id := range ss.Ackmap {
		if ss.Ackmap[id] == Acked {
			ss.Ackmap[id] = notAcked
		}
	}
}
