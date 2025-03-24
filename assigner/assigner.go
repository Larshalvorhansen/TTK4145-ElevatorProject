package assigner

import (
	"Driver-go/config"
	"Driver-go/coordinator"
	"Driver-go/elevator"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
)

type HRAState struct {
	behaviour   string                 `json:"behaviour"`
	floor       int                    `json:"floor"`
	direction   string                 `json:"direction"`
	cabRequests [config.NumFloors]bool `json:"cabRequests"`
}

type HRAInput struct {
	hallRequests [config.NumFloors][2]bool `json:"hallRequests"`
	states       map[string]HRAState       `json:"states"`
}

func CalculateOptimalOrders(sharedState coordinator.SharedState, id int) elevator.Orders {

	stateMap := make(map[string]HRAState)
	for i, v := range sharedState.States {
		if sharedState.Ackmap[i] == coordinator.NotAvailable || v.State.Motorstatus { // removed the additional "... || v.State,Obstructed" for single elevator use
			continue
		} else {
			stateMap[strconv.Itoa(i)] = HRAState{
				behaviour:   v.State.Behaviour.ToString(),
				floor:       v.State.Floor,
				direction:   v.State.Direction.ToString(),
				cabRequests: v.CabRequests,
			}
		}
	}

	// For debugging
	if len(stateMap) == 0 {
		fmt.Println("no elevator states available for assignment!")
		panic("no elevator states available for assignment!")
	}

	hraInput := HRAInput{sharedState.HallRequests, stateMap}

	hraExecutable := ""
	switch runtime.GOOS {
	case "linux":
		hraExecutable = "hall_request_assigner"
	case "windows":
		hraExecutable = "hall_request_assigner.exe"
	default:
		panic("OS not supported")
	}

	jsonBytes, err := json.Marshal(hraInput)
	if err != nil {
		fmt.Println("json.Marshal error: ", err)
		panic("json.Marshal error")
	}

	ret, err := exec.Command("assigner/executables/"+hraExecutable, "-i", "--includeCab", string(jsonBytes)).CombinedOutput()
	if err != nil {
		fmt.Println("exec.Command error: ", err)
		fmt.Println(string(ret))
		panic("exec.Command error")
	}

	output := new(map[string]elevator.Orders)
	err = json.Unmarshal(ret, &output)
	if err != nil {
		fmt.Println("json.Unmarshal error: ", err)
		panic("json.Unmarshal error")
	}

	return (*output)[strconv.Itoa(id)]
}
