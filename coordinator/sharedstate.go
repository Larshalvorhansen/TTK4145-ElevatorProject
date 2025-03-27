package coordinator

import (
	"elevator-project/config"
	"elevator-project/elevator"
	"elevator-project/hardware"
	"elevator-project/network/peers"
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

func (ss *SharedState) addOrder(newOrder hardware.ButtonEvent, localID int) {
	if newOrder.Button == hardware.BT_Cab {
		ss.States[localID].CabRequests[newOrder.Floor] = true
	} else {
		ss.HallRequests[newOrder.Floor][newOrder.Button] = true
	}
}

func (ss *SharedState) addCabCall(newOrder hardware.ButtonEvent, localID int) {
	if newOrder.Button == hardware.BT_Cab {
		ss.States[localID].CabRequests[newOrder.Floor] = true
	}
}

func (ss *SharedState) removeOrder(deliveredOrder hardware.ButtonEvent, localID int) {
	if deliveredOrder.Button == hardware.BT_Cab {
		ss.States[localID].CabRequests[deliveredOrder.Floor] = false
	} else {
		ss.HallRequests[deliveredOrder.Floor][deliveredOrder.Button] = false
	}
}

func (ss *SharedState) updateState(newState elevator.State, localID int) {
	ss.States[localID] = LocalState{
		State:       newState,
		CabRequests: ss.States[localID].CabRequests,
	}
}

func (ss *SharedState) prepareNewState(localID int) {
	ss.Version++
	ss.OriginID = localID
	for localID := range ss.Availability {
		if ss.Availability[localID] == Confirmed {
			ss.Availability[localID] = Unconfirmed
		}
	}
}

func (ss *SharedState) confirm(localID int) {
	ss.Availability[localID] = Confirmed
}

func (ss *SharedState) isFullyConfirmed(localID int) bool {
	if ss.Availability[localID] == Unavailable {
		return false
	}
	for index := range ss.Availability {
		if ss.Availability[index] == Unconfirmed {
			return false
		}
	}
	return true
}

// Ignores Availability, which is node-specific and may vary between peers
func (s1 SharedState) inSyncWith(s2 SharedState) bool {
	return s1.Version == s2.Version &&
		s1.OriginID == s2.OriginID &&
		s1.HallRequests == s2.HallRequests &&
		s1.States == s2.States
}

func (ss *SharedState) setLostPeersUnavailable(peers peers.PeerUpdate) {
	for _, localID := range peers.Lost {
		ss.Availability[localID] = Unavailable
	}
}

func (ss *SharedState) setAllPeersUnavailableExcept(localID int) {
	for elev := range ss.Availability {
		if elev != localID {
			ss.Availability[elev] = Unavailable
		}
	}
}
