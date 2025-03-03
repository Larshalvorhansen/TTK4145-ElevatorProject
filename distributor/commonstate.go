package distributor

import (
	"Driver-go/config"
	"Driver-go/elevator"
	"Driver-go/elevio"
	"Driver-go/network/peers"
	"reflect"
)

type Acknowledgement int

const (
	Acked Acknowledgement = iota
	NotAcked
	NotAvailable
)

type LocalState struct {
	State     elevator.Elevator
	CabOrders [config.NumFloors]bool
}

type CommonState struct {
	ElevatorStates [config.NumElevators]LocalState
	HallOrders     [config.NumFloors][2]bool // request from floors, up or down direction
	Origin         int                       // which elevator is sending the common state
	SeqNumber      int
	AckMap         [config.NumElevators]Acknowledgement
}

/*
addOrder(cs *CommonState, newOrder elevio.ButtonEvent, id int)				DONE
addCabCall(cs *CommonState, newOrder elevio.ButtonEvent, id int)			?HANDELED IN addOrder?
removeOrder(cs *CommonState, deliveredOrder elevio.ButtonEvent, id int)		DONE
updateState(cs *CommonState, newState elevator.State, id int)				DONE
fullyAcked(cs *CommonState, id int) bool									DONE
equals(oldCs CommonState, newCs CommonState) bool 							DONE
makeLostPeersUnavailable(cs *CommonState, peers peers.PeerUpdate)			DONE
makeOthersUnavailable(cs *CommonState, id int) 								DONE
prepNewCs(cs *CommonState, id int)											DONE
*/

func (common_state *CommonState) addOrder(id int, newOrder elevio.ButtonEvent) {
	if newOrder.Button == elevio.BT_Cab {
		common_state.ElevatorStates[id].CabOrders[newOrder.Floor] = true
	} else {
		common_state.HallOrders[newOrder.Floor][int(newOrder.Button)] = true
	}
}

func (common_state *CommonState) removeOrder(id int, deliveredOrder elevio.ButtonEvent) {
	if deliveredOrder.Button == elevio.BT_Cab {
		common_state.ElevatorStates[id].CabOrders[deliveredOrder.Floor] = false
	} else {
		common_state.HallOrders[deliveredOrder.Floor][int(deliveredOrder.Button)] = false
	}
}

func (common_state *CommonState) makeOthersUnavailable(id int) {
	for i := 0; i < config.NumElevators; i++ {
		if i != id {
			common_state.AckMap[i] = NotAvailable
		}
	}
}

func (common_state *CommonState) makeLostPeersUnavailable(peers peers.PeerUpdate) {
	for _, lostID := range peers.Lost {
		common_state.AckMap[int(lostID)] = NotAvailable
	}
}

func (common_state *CommonState) fullyAcknowledged(id int) bool {
	if common_state.AckMap[id] == NotAvailable {
		return false
	}
	for index := range common_state.AckMap {
		if common_state.AckMap[index] == NotAcked {
			return false
		}
	}
	return true
}

// equalCheck checks if two common states are equal with exeption of the AckMap
func (common_state *CommonState) equalCheck(otherCS CommonState) bool {
	common_state.AckMap = [config.NumElevators]Acknowledgement{}
	otherCS.AckMap = [config.NumElevators]Acknowledgement{}
	return reflect.DeepEqual(common_state, otherCS)
}

func (common_state *CommonState) updateElevatorState(id int, newState elevator.Elevator) {
	common_state.ElevatorStates[id].State = newState
}

func (common_state *CommonState) prepNewCS(id int) {
	common_state.Origin = id
	common_state.SeqNumber++
	for i := 0; i < config.NumElevators; i++ {
		if i == id {
			common_state.AckMap[i] = Acked
		} else {
			common_state.AckMap[i] = NotAcked
		}
	}
}
