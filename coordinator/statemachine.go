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
			ss.MakeOthersUnavailable(id)
			fmt.Println("Lost connection to network")
			offline = true

		case peers = <-peersCh:
			ss.MakeOthersUnavailable(id)
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
				ss.PrepNewSs(id)
				ss.AddOrder(newOrder, id)
				ss.Ackmap[id] = Acked
				idle = false

			case deliveredOrder = <-deliveredOrderCh:
				stashType = Remove
				ss.PrepNewSs(id)
				ss.RemoveOrder(deliveredOrder, id)
				ss.Ackmap[id] = Acked
				idle = false

			case newState = <-newStateCh:
				stashType = State
				ss.PrepNewSs(id)
				ss.UpdateState(newState, id)
				ss.Ackmap[id] = Acked
				idle = false

			case arrivedSs := <-networkRx:
				disconnectTimer = time.NewTimer(config.DisconnectTime)
				if arrivedSs.SeqNum > ss.SeqNum || (arrivedSs.Origin > ss.Origin && arrivedSs.SeqNum == ss.SeqNum) {
					ss = arrivedSs
					ss.MakeLostPeersUnavailable(peers)
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
					ss.AddCabCall(newOrder, id)
					confirmedSsCh <- ss
				}

			case deliveredOrder := <-deliveredOrderCh:
				ss.Ackmap[id] = Acked
				ss.RemoveOrder(deliveredOrder, id)
				confirmedSsCh <- ss

			case newState := <-newStateCh:
				if !(newState.Obstructed || newState.Motorstatus) {
					ss.Ackmap[id] = Acked
					ss.UpdateState(newState, id)
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
					ss.MakeLostPeersUnavailable(peers)

				case arrivedSs.FullyAcked(id):
					ss = arrivedSs
					confirmedSsCh <- ss

					switch {
					case ss.Origin != id && stashType != None:
						ss.PrepNewSs(id)

						switch stashType {
						case Add:
							ss.AddOrder(newOrder, id)
							ss.Ackmap[id] = Acked

						case Remove:
							ss.RemoveOrder(deliveredOrder, id)
							ss.Ackmap[id] = Acked

						case State:
							ss.UpdateState(newState, id)
							ss.Ackmap[id] = Acked
						}

					case ss.Origin == id && stashType != None:
						stashType = None
						idle = true

					default:
						idle = true
					}

				case ss.Equals(arrivedSs):
					ss = arrivedSs
					ss.Ackmap[id] = Acked
					ss.MakeLostPeersUnavailable(peers)

				default:
				}
			default:
			}
		}
	}
}
