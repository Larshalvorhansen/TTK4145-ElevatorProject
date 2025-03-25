package coordinator

import (
	"elevator-project/config"
	"elevator-project/elevator"
	"elevator-project/hardware"
	"elevator-project/network/peers"
	"fmt"
	"time"
)

type pendingAction int

const (
	noAction pendingAction = iota
	addOrder
	removeOrder
	updateState
)

func Coordinator(
	confirmedSharedStateCh chan<- SharedState,
	orderDeliveredCh <-chan hardware.ButtonEvent,
	localStateCh <-chan elevator.State,
	sharedStateTxCh chan<- SharedState,
	sharedStateRxCh <-chan SharedState,
	peerUpdateRxCh <-chan peers.PeerUpdate,
	localID int,
) {

	newOrderCh := make(chan hardware.ButtonEvent, config.BufferSize)

	go hardware.PollButtons(newOrderCh)

	var pendingAction pendingAction
	var newOrder hardware.ButtonEvent
	var deliveredOrder hardware.ButtonEvent
	var newState elevator.State
	var peers peers.PeerUpdate
	var ss SharedState

	disconnectTimer := time.NewTimer(config.DisconnectTime)
	sharedStateTicker := time.NewTicker(config.SharedStateBcastInterval)

	idle := true
	offline := false

	for {
		select {
		case <-disconnectTimer.C:
			ss.setAllPeersUnavailableExcept(localID)
			fmt.Printf("[Elevator %d] Lost connection to all peers! Entering offline mode\n", localID)
			offline = true

		case peers = <-peerUpdateRxCh:
			ss.setAllPeersUnavailableExcept(localID)
			idle = false

		case <-sharedStateTicker.C:
			sharedStateTxCh <- ss

		default:
		}

		switch {
		case idle:
			select {
			case newOrder = <-newOrderCh:
				pendingAction = addOrder
				ss.prepareNewState(localID)
				ss.addOrder(newOrder, localID)
				ss.confirm(localID)
				idle = false

			case deliveredOrder = <-orderDeliveredCh:
				pendingAction = removeOrder
				ss.prepareNewState(localID)
				ss.removeOrder(deliveredOrder, localID)
				ss.confirm(localID)
				idle = false

			case newState = <-localStateCh:
				pendingAction = updateState
				ss.prepareNewState(localID)
				ss.updateState(newState, localID)
				ss.confirm(localID)
				idle = false

			case receivedSharedState := <-sharedStateRxCh:
				disconnectTimer = time.NewTimer(config.DisconnectTime)
				if receivedSharedState.Version > ss.Version || (receivedSharedState.OriginID > ss.OriginID && receivedSharedState.Version == ss.Version) {
					ss = receivedSharedState
					ss.setLostPeersUnavailable(peers)
					ss.confirm(localID)
					idle = false
				}

			default:
			}

		case offline:
			select {
			case <-sharedStateRxCh:
				if ss.States[localID].CabRequests == [config.NumFloors]bool{} {
					fmt.Printf("[Elevator %d] Network connection restored. Back online!\n", localID)
					offline = false
				} else {
					ss.Availability[localID] = Unavailable
				}

			case newOrder := <-newOrderCh:
				if !ss.States[localID].State.Motorstatus {
					ss.confirm(localID)
					ss.addCabCall(newOrder, localID)
					confirmedSharedStateCh <- ss
				}

			case deliveredOrder := <-orderDeliveredCh:
				ss.confirm(localID)
				ss.removeOrder(deliveredOrder, localID)
				confirmedSharedStateCh <- ss

			case newState := <-localStateCh:
				if !(newState.Obstructed || newState.Motorstatus) {
					ss.confirm(localID)
					ss.updateState(newState, localID)
					confirmedSharedStateCh <- ss
				}

			default:
			}

		case !idle:
			select {
			case receivedSharedState := <-sharedStateRxCh:
				if receivedSharedState.Version < ss.Version {
					break
				}
				disconnectTimer = time.NewTimer(config.DisconnectTime)

				switch {
				case receivedSharedState.Version > ss.Version || (receivedSharedState.OriginID > ss.OriginID && receivedSharedState.Version == ss.Version):
					ss = receivedSharedState
					ss.confirm(localID)
					ss.setLostPeersUnavailable(peers)

				case receivedSharedState.isFullyConfirmed(localID):
					ss = receivedSharedState
					confirmedSharedStateCh <- ss

					switch {
					case ss.OriginID != localID && pendingAction != noAction:
						ss.prepareNewState(localID)

						switch pendingAction {
						case addOrder:
							ss.addOrder(newOrder, localID)
							ss.confirm(localID)

						case removeOrder:
							ss.removeOrder(deliveredOrder, localID)
							ss.confirm(localID)

						case updateState:
							ss.updateState(newState, localID)
							ss.confirm(localID)
						}

					case ss.OriginID == localID && pendingAction != noAction:
						pendingAction = noAction
						idle = true

					default:
						idle = true
					}

				case ss.inSyncWith(receivedSharedState):
					ss = receivedSharedState
					ss.confirm(localID)
					ss.setLostPeersUnavailable(peers)

				default:
				}
			default:
			}
		}
	}
}
