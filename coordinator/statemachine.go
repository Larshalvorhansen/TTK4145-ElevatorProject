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
	confirmedSsCh chan<- SharedState,
	deliveredOrderCh <-chan hardware.ButtonEvent,
	newStateCh <-chan elevator.State,
	networkTxCh chan<- SharedState,
	networkRxCh <-chan SharedState,
	peersCh <-chan peers.PeerUpdate,
	id int,
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
	intervalTicker := time.NewTicker(config.CoordinatorTick)

	idle := true
	offline := false

	for {
		select {
		case <-disconnectTimer.C:
			ss.setAllPeersUnavailableExcept(id)
			fmt.Println("Lost connection to network")
			offline = true

		case peers = <-peersCh:
			ss.setAllPeersUnavailableExcept(id)
			idle = false

		case <-intervalTicker.C:
			networkTxCh <- ss

		default:
		}

		switch {
		case idle:
			select {
			case newOrder = <-newOrderCh:
				pendingAction = addOrder
				ss.prepareNewState(id)
				ss.addOrder(newOrder, id)
				ss.Availability[id] = Confirmed
				idle = false

			case deliveredOrder = <-deliveredOrderCh:
				pendingAction = removeOrder
				ss.prepareNewState(id)
				ss.removeOrder(deliveredOrder, id)
				ss.Availability[id] = Confirmed
				idle = false

			case newState = <-newStateCh:
				pendingAction = updateState
				ss.prepareNewState(id)
				ss.updateState(newState, id)
				ss.Availability[id] = Confirmed
				idle = false

			case arrivedSs := <-networkRxCh:
				disconnectTimer = time.NewTimer(config.DisconnectTime)
				if arrivedSs.Version > ss.Version || (arrivedSs.OriginID > ss.OriginID && arrivedSs.Version == ss.Version) {
					ss = arrivedSs
					ss.setLostPeersUnavailable(peers)
					ss.Availability[id] = Confirmed
					idle = false
				}

			default:
			}

		case offline:
			select {
			case <-networkRxCh:
				if ss.States[id].CabRequests == [config.NumFloors]bool{} {
					fmt.Println("Regained connection to network")
					offline = false
				} else {
					ss.Availability[id] = Unavailable
				}

			case newOrder := <-newOrderCh:
				if !ss.States[id].State.Motorstatus {
					ss.Availability[id] = Confirmed
					ss.addCabCall(newOrder, id)
					confirmedSsCh <- ss
				}

			case deliveredOrder := <-deliveredOrderCh:
				ss.Availability[id] = Confirmed
				ss.removeOrder(deliveredOrder, id)
				confirmedSsCh <- ss

			case newState := <-newStateCh:
				if !(newState.Obstructed || newState.Motorstatus) {
					ss.Availability[id] = Confirmed
					ss.updateState(newState, id)
					confirmedSsCh <- ss
				}

			default:
			}

		case !idle:
			select {
			case arrivedSs := <-networkRxCh:
				if arrivedSs.Version < ss.Version {
					break
				}
				disconnectTimer = time.NewTimer(config.DisconnectTime)

				switch {
				case arrivedSs.Version > ss.Version || (arrivedSs.OriginID > ss.OriginID && arrivedSs.Version == ss.Version):
					ss = arrivedSs
					ss.Availability[id] = Confirmed
					ss.setLostPeersUnavailable(peers)

				case arrivedSs.isFullyConfirmed(id):
					ss = arrivedSs
					confirmedSsCh <- ss

					switch {
					case ss.OriginID != id && pendingAction != noAction:
						ss.prepareNewState(id)

						switch pendingAction {
						case addOrder:
							ss.addOrder(newOrder, id)
							ss.Availability[id] = Confirmed

						case removeOrder:
							ss.removeOrder(deliveredOrder, id)
							ss.Availability[id] = Confirmed

						case updateState:
							ss.updateState(newState, id)
							ss.Availability[id] = Confirmed
						}

					case ss.OriginID == id && pendingAction != noAction:
						pendingAction = noAction
						idle = true

					default:
						idle = true
					}

				case ss.isEqual(arrivedSs):
					ss = arrivedSs
					ss.Availability[id] = Confirmed
					ss.setLostPeersUnavailable(peers)

				default:
				}
			default:
			}
		}
	}
}
