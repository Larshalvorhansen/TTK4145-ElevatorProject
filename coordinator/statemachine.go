package coordinator

import (
	"Driver-go/config"
	"Driver-go/elevator"
	"Driver-go/hardware"
	"Driver-go/network/peers"
	"log"
	"time"
)

type stashType int

const (
	stachNone stashType = iota
	stachAdd
	stachRemove
	stachState
)

type distributorState int

const (
	stateIdle distributorState = iota
	stateOffline
	stateActive
)

// pendingInfo holds pending actions and data to avoid losing order details.
type pendingInfo struct {
	action   stashType
	order    hardware.ButtonEvent
	newState elevator.State
}

// Distributor handles order distribution and network updates in separate states.
func Distributor(
	confirmedCsC chan<- SharedState,
	deliveredOrderC <-chan hardware.ButtonEvent,
	newStateC <-chan elevator.State,
	networkTx chan<- SharedState,
	networkRx <-chan SharedState,
	peersC <-chan peers.PeerUpdate,
	id int,
) {
	newOrderC := make(chan hardware.ButtonEvent, config.BufferSize)
	go hardware.PollButtons(newOrderC)

	var pending pendingInfo
	var commonState SharedState
	var currentState distributorState = stateIdle

	disconnectTimer := time.NewTimer(config.DisconnectTime)
	intervalTicker := time.NewTicker(config.DistributorTick)

	for {
		select {
		case <-disconnectTimer.C:
			commonState.makeOthersUnavailable(id)
			log.Println("Network connection lost")
			currentState = stateOffline

		case updatePeers := <-peersC:
			commonState.makeOthersUnavailable(id)
			if currentState == stateIdle {
				currentState = stateActive
			}
			commonState.makeLostPeersUnavailable(updatePeers)

		case <-intervalTicker.C:
			networkTx <- commonState
		}

		switch currentState {
		case stateIdle:
			currentState = handleIdleState(
				&commonState,
				id,
				&pending,
				newOrderC,
				deliveredOrderC,
				newStateC,
				networkRx,
				disconnectTimer,
			)
		case stateOffline:
			currentState = handleOfflineState(
				&commonState,
				id,
				confirmedCsC,
				newOrderC,
				deliveredOrderC,
				newStateC,
				networkRx,
			)
		case stateActive:
			currentState = handleActiveState(
				&commonState,
				id,
				&pending,
				confirmedCsC,
				newOrderC,
				deliveredOrderC,
				newStateC,
				networkRx,
				disconnectTimer,
			)
		}
	}
}

func handleIdleState(
	ss *SharedState,
	id int,
	pending *pendingInfo,
	newOrderC <-chan hardware.ButtonEvent,
	deliveredOrderC <-chan hardware.ButtonEvent,
	newStateC <-chan elevator.State,
	networkRx <-chan SharedState,
	disconnectTimer *time.Timer,
) distributorState {
	select {
	case newOrder := <-newOrderC:
		pending.action = stachAdd
		pending.order = newOrder
		ss.prepNewCs(id)
		ss.addOrder(newOrder, id)
		ss.Ackmap[id] = Acked
		return stateActive

	case deliveredOrder := <-deliveredOrderC:
		pending.action = stachRemove
		pending.order = deliveredOrder
		ss.prepNewCs(id)
		ss.removeOrder(deliveredOrder, id)
		ss.Ackmap[id] = Acked
		return stateActive

	case newSt := <-newStateC:
		pending.action = stachState
		pending.newState = newSt
		ss.prepNewCs(id)
		ss.updateState(newSt, id)
		ss.Ackmap[id] = Acked
		return stateActive

	case incomingState := <-networkRx:
		disconnectTimer.Reset(config.DisconnectTime)
		if isNewer(incomingState, *ss) {
			*ss = incomingState
			ss.Ackmap[id] = Acked
			return stateActive
		}
	default:
	}
	return stateIdle
}

func handleOfflineState(
	ss *SharedState,
	id int,
	confirmedCsC chan<- SharedState,
	newOrderC <-chan hardware.ButtonEvent,
	deliveredOrderC <-chan hardware.ButtonEvent,
	newStateC <-chan elevator.State,
	networkRx <-chan SharedState,
) distributorState {
	select {
	case <-networkRx:
		if ss.States[id].CabRequests == [config.NumFloors]bool{} {
			log.Println("Network connection regained")
			return stateIdle
		}
		ss.Ackmap[id] = NotAvailable

	case newOrder := <-newOrderC:
		if !ss.States[id].State.MotorStatus {
			ss.Ackmap[id] = Acked
			ss.addCabCall(newOrder, id)
			confirmedCsC <- *ss
		}

	case deliveredOrder := <-deliveredOrderC:
		ss.Ackmap[id] = Acked
		ss.removeOrder(deliveredOrder, id)
		confirmedCsC <- *ss

	case newSt := <-newStateC:
		if !(newSt.Obstructed || newSt.MotorStatus) {
			ss.Ackmap[id] = Acked
			ss.updateState(newSt, id)
			confirmedCsC <- *ss
		}
	default:
	}
	return stateOffline
}

func handleActiveState(
	ss *SharedState,
	id int,
	pending *pendingInfo,
	confirmedCsC chan<- SharedState,
	newOrderC <-chan hardware.ButtonEvent,
	deliveredOrderC <-chan hardware.ButtonEvent,
	newStateC <-chan elevator.State,
	networkRx <-chan SharedState,
	disconnectTimer *time.Timer,
) distributorState {
	select {
	case incomingState := <-networkRx:
		if incomingState.SeqNum < ss.SeqNum {
			break
		}
		disconnectTimer.Reset(config.DisconnectTime)

		switch {
		case isNewer(incomingState, *ss):
			*ss = incomingState
			ss.Ackmap[id] = Acked

		case incomingState.fullyAcked(id):
			*ss = incomingState
			confirmedCsC <- *ss
			if ss.Origin != id && pending.action != stachNone {
				ss.prepNewCs(id)
				switch pending.action {
				case stachAdd:
					ss.addOrder(pending.order, id)
				case stachRemove:
					ss.removeOrder(pending.order, id)
				case stachState:
					ss.updateState(pending.newState, id)
				}
				ss.Ackmap[id] = Acked
			} else {
				pending.action = stachNone
				return stateIdle
			}

		case ss.equals(incomingState):
			*ss = incomingState
			ss.Ackmap[id] = Acked
		}
	default:
	}
	return stateActive
}
