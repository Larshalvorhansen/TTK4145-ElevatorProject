package distributor

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
	confirmedCsC chan<- CommonState,
	deliveredOrderC <-chan hardware.ButtonEvent,
	newStateC <-chan elevator.State,
	networkTx chan<- CommonState,
	networkRx <-chan CommonState,
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
	var cs CommonState

	disconnectTimer := time.NewTimer(config.DisconnectTime)
	intervalTicker := time.NewTicker(config.Interval)

	idle := true
	offline := false

	for {
		select {
		case <-disconnectTimer.C:
			cs.makeOthersUnavailable(id)
			fmt.Println("Lost connection to network")
			offline = true

		case peers = <-peersC:
			cs.makeOthersUnavailable(id)
			idle = false

		case <-intervalTicker.C:
			networkTx <- cs

		default:
		}

		switch {
		case idle:
			select {
			case newOrder = <-newOrderC:
				stashType = Add
				cs.prepNewCs(id)
				cs.addOrder(newOrder, id)
				cs.Ackmap[id] = Acked
				idle = false

			case deliveredOrder = <-deliveredOrderC:
				stashType = Remove
				cs.prepNewCs(id)
				cs.removeOrder(deliveredOrder, id)
				cs.Ackmap[id] = Acked
				idle = false

			case newState = <-newStateC:
				stashType = State
				cs.prepNewCs(id)
				cs.updateState(newState, id)
				cs.Ackmap[id] = Acked
				idle = false

			case arrivedCs := <-networkRx:
				disconnectTimer = time.NewTimer(config.DisconnectTime)
				if arrivedCs.SeqNum > cs.SeqNum || (arrivedCs.Origin > cs.Origin && arrivedCs.SeqNum == cs.SeqNum) {
					cs = arrivedCs
					cs.makeLostPeersUnavailable(peers)
					cs.Ackmap[id] = Acked
					idle = false
				}

			default:
			}

		case offline:
			select {
			case <-networkRx:
				if cs.States[id].CabRequests == [config.NumFloors]bool{} {
					fmt.Println("Regained connection to network")
					offline = false
				} else {
					cs.Ackmap[id] = NotAvailable
				}

			case newOrder := <-newOrderC:
				if !cs.States[id].State.Motorstatus {
					cs.Ackmap[id] = Acked
					cs.addCabCall(newOrder, id)
					confirmedCsC <- cs
				}

			case deliveredOrder := <-deliveredOrderC:
				cs.Ackmap[id] = Acked
				cs.removeOrder(deliveredOrder, id)
				confirmedCsC <- cs

			case newState := <-newStateC:
				if !(newState.Obstructed || newState.Motorstatus) {
					cs.Ackmap[id] = Acked
					cs.updateState(newState, id)
					confirmedCsC <- cs
				}

			default:
			}

		case !idle:
			select {
			case arrivedCs := <-networkRx:
				if arrivedCs.SeqNum < cs.SeqNum {
					break
				}
				disconnectTimer = time.NewTimer(config.DisconnectTime)

				switch {
				case arrivedCs.SeqNum > cs.SeqNum || (arrivedCs.Origin > cs.Origin && arrivedCs.SeqNum == cs.SeqNum):
					cs = arrivedCs
					cs.Ackmap[id] = Acked
					cs.makeLostPeersUnavailable(peers)

				case arrivedCs.fullyAcked(id):
					cs = arrivedCs
					confirmedCsC <- cs

					switch {
					case cs.Origin != id && stashType != None:
						cs.prepNewCs(id)

						switch stashType {
						case Add:
							cs.addOrder(newOrder, id)
							cs.Ackmap[id] = Acked

						case Remove:
							cs.removeOrder(deliveredOrder, id)
							cs.Ackmap[id] = Acked

						case State:
							cs.updateState(newState, id)
							cs.Ackmap[id] = Acked
						}

					case cs.Origin == id && stashType != None:
						stashType = None
						idle = true

					default:
						idle = true
					}

				case cs.equals(arrivedCs):
					cs = arrivedCs
					cs.Ackmap[id] = Acked
					cs.makeLostPeersUnavailable(peers)

				default:
				}
			default:
			}
		}
	}
}
