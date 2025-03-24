package coordinator

import (
	"Driver-go/config"
	"Driver-go/elevator"
	"Driver-go/hardware"
	"Driver-go/network/peers"
	"fmt"
	"time"
)

type StashType int

const (
	None StashType = iota
	Add
	Remove
	State
)

func Distributor(
	confirmedSsCh chan<- SharedState,
	deliveredOrderCh <-chan hardware.ButtonEvent,
	newStateCh <-chan elevator.State,
	networkTx chan<- SharedState,
	networkRx <-chan SharedState,
	peersCh <-chan peers.PeerUpdate,
	id int,
) {

	newOrderCh := make(chan hardware.ButtonEvent, config.BufferSize)

	go hardware.PollButtons(newOrderCh)

	var stashType StashType
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
			ss.makeOthersUnavailable(id)
			fmt.Println("Lost connection to network")
			offline = true

		case peers = <-peersCh:
			ss.makeOthersUnavailable(id)
			idle = false

		case <-intervalTicker.C:
			networkTx <- ss

		default:
		}

		switch {
		case idle:
			select {
			case newOrder = <-newOrderCh:
				stashType = Add
				ss.prepNewSs(id)
				ss.addOrder(newOrder, id)
				ss.Ackmap[id] = Acked
				idle = false

			case deliveredOrder = <-deliveredOrderCh:
				stashType = Remove
				ss.prepNewSs(id)
				ss.removeOrder(deliveredOrder, id)
				ss.Ackmap[id] = Acked
				idle = false

			case newState = <-newStateCh:
				stashType = State
				ss.prepNewSs(id)
				ss.updateState(newState, id)
				ss.Ackmap[id] = Acked
				idle = false

			case arrivedSs := <-networkRx:
				disconnectTimer = time.NewTimer(config.DisconnectTime)
				if arrivedSs.SeqNum > ss.SeqNum || (arrivedSs.Origin > ss.Origin && arrivedSs.SeqNum == ss.SeqNum) {
					ss = arrivedSs
					ss.makeLostPeersUnavailable(peers)
					ss.Ackmap[id] = Acked
					idle = false
				}

			default:
			}

		case offline:
			select {
			case <-networkRx:
				if ss.States[id].CabRequests == [config.NumFloors]bool{} {
					fmt.Println("Regained connection to network")
					offline = false
				} else {
					ss.Ackmap[id] = NotAvailable
				}

			case newOrder := <-newOrderCh:
				if !ss.States[id].State.Motorstatus {
					ss.Ackmap[id] = Acked
					ss.addCabCall(newOrder, id)
					confirmedSsCh <- ss
				}

			case deliveredOrder := <-deliveredOrderCh:
				ss.Ackmap[id] = Acked
				ss.removeOrder(deliveredOrder, id)
				confirmedSsCh <- ss

			case newState := <-newStateCh:
				if !(newState.Obstructed || newState.Motorstatus) {
					ss.Ackmap[id] = Acked
					ss.updateState(newState, id)
					confirmedSsCh <- ss
				}

			default:
			}

		case !idle:
			select {
			case arrivedSs := <-networkRx:
				if arrivedSs.SeqNum < ss.SeqNum {
					break
				}
				disconnectTimer = time.NewTimer(config.DisconnectTime)

				switch {
				case arrivedSs.SeqNum > ss.SeqNum || (arrivedSs.Origin > ss.Origin && arrivedSs.SeqNum == ss.SeqNum):
					ss = arrivedSs
					ss.Ackmap[id] = Acked
					ss.makeLostPeersUnavailable(peers)

				case arrivedSs.fullyAcked(id):
					ss = arrivedSs
					confirmedSsCh <- ss

					switch {
					case ss.Origin != id && stashType != None:
						ss.prepNewSs(id)

						switch stashType {
						case Add:
							ss.addOrder(newOrder, id)
							ss.Ackmap[id] = Acked

						case Remove:
							ss.removeOrder(deliveredOrder, id)
							ss.Ackmap[id] = Acked

						case State:
							ss.updateState(newState, id)
							ss.Ackmap[id] = Acked
						}

					case ss.Origin == id && stashType != None:
						stashType = None
						idle = true

					default:
						idle = true
					}

				case ss.equals(arrivedSs):
					ss = arrivedSs
					ss.Ackmap[id] = Acked
					ss.makeLostPeersUnavailable(peers)

				default:
				}
			default:
			}
		}
	}
}
