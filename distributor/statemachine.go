package distributor

import (
	"Driver-go/config"
	"Driver-go/elevator"
	"Driver-go/hardware"
	"Driver-go/network/peers"
	"fmt"
	"time"
)

type ActionType int

const (
	None ActionType = iota
	Add
	Remove
	State
)

func Distributor(
	confirmedSVCh chan<- SystemView,
	deliveredOrderCh <-chan hardware.ButtonEvent,
	newElevStateCh <-chan elevator.State,
	networkTxSVCh chan<- SystemView,
	networkRxSVCh <-chan SystemView,
	peersCh <-chan peers.PeerUpdate,
	id int,
) {

	newOrderCh := make(chan hardware.ButtonEvent, config.BufferSize)

	go hardware.PollButtons(newOrderCh)

	var actionType ActionType
	var newOrder hardware.ButtonEvent
	var deliveredOrder hardware.ButtonEvent
	var newElevState elevator.State
	var peers peers.PeerUpdate
	var sv SystemView

	disconnectTimer := time.NewTimer(config.DisconnectTime)
	intervalTicker := time.NewTicker(config.Interval)

	idle := true
	offline := false

	for {
		select {
		case <-disconnectTimer.C:
			sv.makeOthersUnavailable(id)
			fmt.Println("Lost connection to network")
			offline = true

		case peers = <-peersCh:
			sv.makeOthersUnavailable(id)
			idle = false

		case <-intervalTicker.C:
			networkTxSVCh <- sv

		default:
		}

		switch {
		case idle:
			select {
			case newOrder = <-newOrderCh:
				actionType = Add
				sv.prepNewSV(id)
				sv.addOrder(newOrder, id)
				sv.Ackmap[id] = Acked
				idle = false

			case deliveredOrder = <-deliveredOrderCh:
				actionType = Remove
				sv.prepNewSV(id)
				sv.removeOrder(deliveredOrder, id)
				sv.Ackmap[id] = Acked
				idle = false

			case newElevState = <-newElevStateCh:
				actionType = State
				sv.prepNewSV(id)
				sv.updateState(newElevState, id)
				sv.Ackmap[id] = Acked
				idle = false

			case arrivedSV := <-networkRxSVCh:
				disconnectTimer = time.NewTimer(config.DisconnectTime)
				if arrivedSV.SeqNum > sv.SeqNum || (arrivedSV.Origin > sv.Origin && arrivedSV.SeqNum == sv.SeqNum) {
					sv = arrivedSV
					sv.makeLostPeersUnavailable(peers)
					sv.Ackmap[id] = Acked
					idle = false
				}

			default:
			}

		case offline:
			select {
			case <-networkRxSVCh:
				if sv.ElevatorStates[id].CabRequests == [config.NumFloors]bool{} {
					fmt.Println("Regained connection to network")
					offline = false
				} else {
					sv.Ackmap[id] = NotAvailable
				}

			case newOrder := <-newOrderCh:
				if !sv.ElevatorStates[id].State.Motorstatus {
					sv.Ackmap[id] = Acked
					sv.addCabCall(newOrder, id)
					confirmedSVCh <- sv
				}

			case deliveredOrder := <-deliveredOrderCh:
				sv.Ackmap[id] = Acked
				sv.removeOrder(deliveredOrder, id)
				confirmedSVCh <- sv

			case newElevState := <-newElevStateCh:
				if !(newElevState.Obstructed || newElevState.Motorstatus) {
					sv.Ackmap[id] = Acked
					sv.updateState(newElevState, id)
					confirmedSVCh <- sv
				}

			default:
			}

		case !idle:
			select {
			case arrivedSV := <-networkRxSVCh:
				if arrivedSV.SeqNum < sv.SeqNum {
					break
				}
				disconnectTimer = time.NewTimer(config.DisconnectTime)

				switch {
				case arrivedSV.SeqNum > sv.SeqNum || (arrivedSV.Origin > sv.Origin && arrivedSV.SeqNum == sv.SeqNum):
					sv = arrivedSV
					sv.Ackmap[id] = Acked
					sv.makeLostPeersUnavailable(peers)

				case arrivedSV.fullyAcked(id):
					sv = arrivedSV
					confirmedSVCh <- sv

					switch {
					case sv.Origin != id && actionType != None:
						sv.prepNewSV(id)

						switch actionType {
						case Add:
							sv.addOrder(newOrder, id)
							sv.Ackmap[id] = Acked

						case Remove:
							sv.removeOrder(deliveredOrder, id)
							sv.Ackmap[id] = Acked

						case State:
							sv.updateState(newElevState, id)
							sv.Ackmap[id] = Acked
						}

					case sv.Origin == id && actionType != None:
						actionType = None
						idle = true

					default:
						idle = true
					}

				case sv.equals(arrivedSV):
					sv = arrivedSV
					sv.Ackmap[id] = Acked
					sv.makeLostPeersUnavailable(peers)

				default:
				}
			default:
			}
		}
	}
}
