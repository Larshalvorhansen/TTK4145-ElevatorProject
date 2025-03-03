package distributor

import (
	"Driver-go/config"
	"Driver-go/elevator"
	"Driver-go/elevio"
	"Driver-go/network/peers"
	"reflect"
	"strconv"
)

type Acknowledgement int

const (
	Acked Acknowledgement = iota
	NotAcked
	NotAvailable
)

type LocalState struct {
	State     elevator.State
	CabOrders [config.NumFloors]bool
}

type CommonState struct {
	ElevatorStates [config.NumElevators]LocalState
	HallOrders     [config.NumFloors][2]bool // request from floors, up or down direction
	Origin         int                       // which elevator is sending the common state
	SeqNumber      int
	AckMap         [config.NumElevators]Acknowledgement
}

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
		intLostID, error := strconv.Atoi(lostID)
		if error == nil {
			common_state.AckMap[intLostID] = NotAvailable
		}
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

func (common_state *CommonState) updateElevatorState(id int, newState elevator.State) {
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
