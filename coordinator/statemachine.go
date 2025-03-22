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

		case peers = <-peersC:
			ss.makeOthersUnavailable(id)
			idle = false

		case <-intervalTicker.C:
			networkTx <- ss

		default:
		}

		switch {
		case idle:
			select {
			case newOrder = <-newOrderC:
				stashType = Add
				ss.prepNewCs(id)
				ss.addOrder(newOrder, id)
				ss.Ackmap[id] = Acked
				idle = false

			case deliveredOrder = <-deliveredOrderC:
				stashType = Remove
				ss.prepNewCs(id)
				ss.removeOrder(deliveredOrder, id)
				ss.Ackmap[id] = Acked
				idle = false

			case newState = <-newStateC:
				stashType = State
				ss.prepNewCs(id)
				ss.updateState(newState, id)
				ss.Ackmap[id] = Acked
				idle = false

			case arrivedCs := <-networkRx:
				disconnectTimer = time.NewTimer(config.DisconnectTime)
				if arrivedCs.SeqNum > ss.SeqNum || (arrivedCs.Origin > ss.Origin && arrivedCs.SeqNum == ss.SeqNum) {
					ss = arrivedCs
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

			case newOrder := <-newOrderC:
				if !ss.States[id].State.Motorstatus {
					ss.Ackmap[id] = Acked
					ss.addCabCall(newOrder, id)
					confirmedCsC <- ss
				}

			case deliveredOrder := <-deliveredOrderC:
				ss.Ackmap[id] = Acked
				ss.removeOrder(deliveredOrder, id)
				confirmedCsC <- ss

			case newState := <-newStateC:
				if !(newState.Obstructed || newState.Motorstatus) {
					ss.Ackmap[id] = Acked
					ss.updateState(newState, id)
					confirmedCsC <- ss
				}

			default:
			}

		case !idle:
			select {
			case arrivedCs := <-networkRx:
				if arrivedCs.SeqNum < ss.SeqNum {
					break
				}
				disconnectTimer = time.NewTimer(config.DisconnectTime)

				switch {
				case arrivedCs.SeqNum > ss.SeqNum || (arrivedCs.Origin > ss.Origin && arrivedCs.SeqNum == ss.SeqNum):
					ss = arrivedCs
					ss.Ackmap[id] = Acked
					ss.makeLostPeersUnavailable(peers)

				case arrivedCs.fullyAcked(id):
					ss = arrivedCs
					confirmedCsC <- ss

					switch {
					case ss.Origin != id && stashType != None:
						ss.prepNewCs(id)

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

				case ss.equals(arrivedCs):
					ss = arrivedCs
					ss.Ackmap[id] = Acked
					ss.makeLostPeersUnavailable(peers)

				default:
				}
			default:
			}
		}
	}
}
