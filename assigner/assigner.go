// The following implementation is based on the example provided by TTK4145's project resources:
// https://github.com/TTK4145/Project-resources/blob/master/cost_fns/usage_examples/example.go
// Modifications were made to integrate it into the current project's codebase and requirements.

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

// Struct members must be public in order to be accessible by json.Marshal/.Unmarshal
// This means they must start with a capital letter, so we need to use field renaming struct tags to make them camelCase

type HRAElevState struct {
	Behaviour   string                 `json:"behaviour"`
	Floor       int                    `json:"floor"`
	Direction   string                 `json:"direction"`
	CabRequests [config.NumFloors]bool `json:"cabRequests"`
}

type HRAInput struct {
	HallRequests [config.NumFloors][2]bool `json:"hallRequests"`
	States       map[string]HRAElevState   `json:"states"`
}

func AssignOrders(ss coordinator.SharedState, id int) elevator.Orders {

	hraExecutable := ""
	switch runtime.GOOS {
	case "linux":
		hraExecutable = "hall_request_assigner"
	case "darwin":
		hraExecutable = "hall_request_assigner_mac"
	case "windows":
		hraExecutable = "hall_request_assigner.exe"
	default:
		panic("OS not supported")
	}

	stateMap := make(map[string]HRAElevState)
	for i, v := range ss.States {
		if ss.Ackmap[i] == coordinator.NotAvailable || v.State.Motorstatus || v.State.Obstructed { // For single elevator use, comment out the last two conditions
			continue
		} else {
			stateMap[strconv.Itoa(i)] = HRAElevState{
				Behaviour:   v.State.Behaviour.ToString(),
				Floor:       v.State.Floor,
				Direction:   v.State.Direction.ToString(),
				CabRequests: v.CabRequests,
			}
		}
	}

	// For debugging
	if len(stateMap) == 0 {
		panic("no elevator states available for assignment!")
	}

	input := HRAInput{ss.HallRequests, stateMap}

	jsonBytes, err := json.Marshal(input)
	if err != nil {
		panic(fmt.Sprintf("json.Marshal error: %v", err))
	}

	ret, err := exec.Command("assigner/executables/"+hraExecutable, "-i", "--includeCab", string(jsonBytes)).CombinedOutput()
	if err != nil {
		panic(fmt.Sprintf("exec.Command error: %v, output: %s", err, string(ret)))
	}

	output := new(map[string]elevator.Orders)
	err = json.Unmarshal(ret, &output)
	if err != nil {
		panic(fmt.Sprintf("json.Unmarshal error: %v", err))
	}

	return (*output)[strconv.Itoa(id)]
}
