package distributor

import (
	"Driver-go/config"
	"Driver-go/elevio"
	"Driver-go/elevator"
	"Driver-go/network/peers"
)

type Local_State{
	State

}

/*
addOrder(cs *CommonState, newOrder elevio.ButtonEvent, id int)
addCabCall(cs *CommonState, newOrder elevio.ButtonEvent, id int)
removeOrder(cs *CommonState, deliveredOrder elevio.ButtonEvent, id int)
updateState(cs *CommonState, newState elevator.State, id int)
fullyAcked(cs *CommonState, id int) bool
equals(oldCs CommonState, newCs CommonState) bool
makeLostPeersUnavailable(cs *CommonState, peers peers.PeerUpdate)
makeOthersUnavailable(cs *CommonState, id int)
prepNewCs(cs *CommonState, id int)
*/