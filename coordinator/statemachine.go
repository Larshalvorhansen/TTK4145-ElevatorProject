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
			fmt.Println(fmt.Sprintf("[Elevator %d] Lost connection to all peers! Entering offline mode", localID))
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
				ss.Availability[localID] = Confirmed
				idle = false

			case deliveredOrder = <-orderDeliveredCh:
				pendingAction = removeOrder
				ss.prepareNewState(localID)
				ss.removeOrder(deliveredOrder, localID)
				ss.Availability[localID] = Confirmed
				idle = false

			case newState = <-localStateCh:
				pendingAction = updateState
				ss.prepareNewState(localID)
				ss.updateState(newState, localID)
				ss.Availability[localID] = Confirmed
				idle = false

			case arrivedSs := <-sharedStateRxCh:
				disconnectTimer = time.NewTimer(config.DisconnectTime)
				if arrivedSs.Version > ss.Version || (arrivedSs.OriginID > ss.OriginID && arrivedSs.Version == ss.Version) {
					ss = arrivedSs
					ss.setLostPeersUnavailable(peers)
					ss.Availability[localID] = Confirmed
					idle = false
				}

			default:
			}

		case offline:
			select {
			case <-sharedStateRxCh:
				if ss.States[localID].CabRequests == [config.NumFloors]bool{} {
					fmt.Println(fmt.Sprintf("[Elevator %d] Network connection restored. Back online!", localID))
					offline = false
				} else {
					ss.Availability[localID] = Unavailable
				}

			case newOrder := <-newOrderCh:
				if !ss.States[localID].State.Motorstatus {
					ss.Availability[localID] = Confirmed
					ss.addCabCall(newOrder, localID)
					confirmedSharedStateCh <- ss
				}

			case deliveredOrder := <-orderDeliveredCh:
				ss.Availability[localID] = Confirmed
				ss.removeOrder(deliveredOrder, localID)
				confirmedSharedStateCh <- ss

			case newState := <-localStateCh:
				if !(newState.Obstructed || newState.Motorstatus) {
					ss.Availability[localID] = Confirmed
					ss.updateState(newState, localID)
					confirmedSharedStateCh <- ss
				}

			default:
			}

		case !idle:
			select {
			case arrivedSs := <-sharedStateRxCh:
				if arrivedSs.Version < ss.Version {
					break
				}
				disconnectTimer = time.NewTimer(config.DisconnectTime)

				switch {
				case arrivedSs.Version > ss.Version || (arrivedSs.OriginID > ss.OriginID && arrivedSs.Version == ss.Version):
					ss = arrivedSs
					ss.Availability[localID] = Confirmed
					ss.setLostPeersUnavailable(peers)

				case arrivedSs.isFullyConfirmed(localID):
					ss = arrivedSs
					confirmedSharedStateCh <- ss

					switch {
					case ss.OriginID != localID && pendingAction != noAction:
						ss.prepareNewState(localID)

						switch pendingAction {
						case addOrder:
							ss.addOrder(newOrder, localID)
							ss.Availability[localID] = Confirmed

						case removeOrder:
							ss.removeOrder(deliveredOrder, localID)
							ss.Availability[localID] = Confirmed

						case updateState:
							ss.updateState(newState, localID)
							ss.Availability[localID] = Confirmed
						}

					case ss.OriginID == localID && pendingAction != noAction:
						pendingAction = noAction
						idle = true

					default:
						idle = true
					}

				case ss.isEqual(arrivedSs):
					ss = arrivedSs
					ss.Availability[localID] = Confirmed
					ss.setLostPeersUnavailable(peers)

				default:
				}
			default:
			}
		}
	}
}
