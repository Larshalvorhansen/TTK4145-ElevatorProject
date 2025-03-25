package coordinator

import (
	"elevator-project/config"
	"elevator-project/elevator"
	"elevator-project/hardware"
	"elevator-project/network/peers"
	"reflect"
)

type ConfirmationStatus int

const (
	Unconfirmed ConfirmationStatus = iota
	Confirmed
	Unavailable
)

type LocalState struct {
	State       elevator.State
	CabRequests [config.NumFloors]bool
}

type SharedState struct {
	Version      int
	OriginID     int
	Availability [config.NumElevators]ConfirmationStatus
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

func (ss *SharedState) isFullyConfirmed(id int) bool {
	if ss.Availability[id] == Unavailable {
		return false
	}
	for index := range ss.Availability {
		if ss.Availability[index] == Unconfirmed {
			return false
		}
	}
	return true
}

func (oldSs SharedState) isEqual(newSs SharedState) bool {
	oldSs.Availability = [config.NumElevators]ConfirmationStatus{}
	newSs.Availability = [config.NumElevators]ConfirmationStatus{}
	return reflect.DeepEqual(oldSs, newSs)
}

func (ss *SharedState) setLostPeersUnavailable(peers peers.PeerUpdate) {
	for _, id := range peers.Lost {
		ss.Availability[id] = Unavailable
	}
}

func (ss *SharedState) setAllPeersUnavailableExcept(id int) {
	for elev := range ss.Availability {
		if elev != id {
			ss.Availability[elev] = Unavailable
		}
	}
}

func (ss *SharedState) prepareNewState(id int) {
	ss.Version++
	ss.OriginID = id
	for id := range ss.Availability {
		if ss.Availability[id] == Confirmed {
			ss.Availability[id] = Unconfirmed
		}
	}
}
