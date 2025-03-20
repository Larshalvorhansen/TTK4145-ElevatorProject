// TODO: check iota constants, if they are exported capitalize firt letter, if not, keep it lowercase in first letter

package coordinator

import (
	"Driver-go/config"
	"Driver-go/elevator"
	"Driver-go/hardware"
	"Driver-go/network/peers"
	"reflect"
)

// AckStatus indicates the acknowledgment state for each elevator.
type AckStatus int

const (
	NotAcked AckStatus = iota
	Acked
	NotAvailable
)

// LocalState holds the elevator's current State and its cab requests.
type LocalState struct {
	State       elevator.State
	CabRequests [config.NumFloors]bool
}

// SharedState represents the shared state among all elevators in the system.
type SharedState struct {
	SeqNum       int
	Origin       int
	Ackmap       [config.NumElevators]AckStatus
	HallRequests [config.NumFloors][2]bool
	States       [config.NumElevators]LocalState
}

// isNewer checks if 'a' is strictly newer than 'b' based on SeqNum and Origin.
func isNewer(a, b SharedState) bool {
	if a.SeqNum > b.SeqNum {
		return true
	}
	if a.SeqNum == b.SeqNum && a.Origin > b.Origin {
		return true
	}
	return false
}

// addOrder places a new hall or cab request into the shared state.
func (ss *SharedState) addOrder(newOrder hardware.ButtonEvent, elevatorID int) {
	if newOrder.Button == hardware.BT_Cab {
		ss.States[elevatorID].CabRequests[newOrder.Floor] = true
	} else {
		ss.HallRequests[newOrder.Floor][newOrder.Button] = true
	}
}

// addCabCall specifically adds a cab request, used when offline or in fallback logic.
func (ss *SharedState) addCabCall(newOrder hardware.ButtonEvent, elevatorID int) {
	if newOrder.Button == hardware.BT_Cab {
		ss.States[elevatorID].CabRequests[newOrder.Floor] = true
	}
}

// removeOrder clears a delivered order (hall or cab) from the shared state.
func (ss *SharedState) removeOrder(deliveredOrder hardware.ButtonEvent, elevatorID int) {
	if deliveredOrder.Button == hardware.BT_Cab {
		ss.States[elevatorID].CabRequests[deliveredOrder.Floor] = false
	} else {
		ss.HallRequests[deliveredOrder.Floor][deliveredOrder.Button] = false
	}
}

// updateState updates the internal elevator state while preserving cab requests.
func (ss *SharedState) updateState(newState elevator.State, elevatorID int) {
	ss.States[elevatorID] = LocalState{
		State:       newState,
		CabRequests: ss.States[elevatorID].CabRequests,
	}
}

// fullyAcked returns true if all elevators have acknowledged the latest update.
func (ss *SharedState) fullyAcked(elevatorID int) bool {
	// If we marked it NotAvailable, we consider it not part of the ack check
	if ss.Ackmap[elevatorID] == NotAvailable {
		return false
	}
	for i := range ss.Ackmap {
		if ss.Ackmap[i] == NotAcked {
			return false
		}
	}
	return true
}

// equals checks whether two CommonStates differ only by their Ackmap.
// It zeroes out each Ackmap, then compares the rest with DeepEqual.
func (oldCs SharedState) equals(newCs SharedState) bool {
	oldCs.Ackmap = [config.NumElevators]AckStatus{}
	newCs.Ackmap = [config.NumElevators]AckStatus{}
	return reflect.DeepEqual(oldCs, newCs)
}

// makeLostPeersUnavailable marks all lost peers in the Ackmap as NotAvailable.
func (ss *SharedState) makeLostPeersUnavailable(update peers.PeerUpdate) {
	for _, lostID := range update.Lost {
		ss.Ackmap[lostID] = NotAvailable
	}
}

// makeOthersUnavailable marks all elevators except the given one as NotAvailable.
func (ss *SharedState) makeOthersUnavailable(except int) {
	for i := range ss.Ackmap {
		if i != except {
			ss.Ackmap[i] = NotAvailable
		}
	}
}

// prepNewCs increments the SeqNum, sets Origin, and sets Acked -> NotAcked on others.
func (ss *SharedState) prepNewCs(elevatorID int) {
	ss.SeqNum++
	ss.Origin = elevatorID
	for i := range ss.Ackmap {
		if ss.Ackmap[i] == Acked {
			ss.Ackmap[i] = NotAcked
		}
	}
}
