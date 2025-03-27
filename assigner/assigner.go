// TODO: Change func DistributeElevatorOrders to AssignOrders or Assigner (like in coordinator is called Coordinator)? Change elev to elevator?

// The following implementation is based on the example provided by TTK4145's project resources:
// https://github.com/TTK4145/Project-resources/blob/master/cost_fns/usage_examples/example.go
// Modifications were made to integrate it into the current project's codebase and requirements.

package assigner

import (
	"elevator-project/config"
	"elevator-project/coordinator"
	"elevator-project/elevator"
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

func DistributeElevatorOrders(ss coordinator.SharedState, localID int) elevator.Orders {

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

	for id, elev := range ss.States {
		isUnavailable := ss.Availability[id] == coordinator.Unavailable
		hasMotorIssue := elev.State.Motorstatus
		hasBlockingObstruction := elev.State.Obstructed && elev.State.Behaviour == elevator.DoorOpen

		// For single elevator use, comment out the two last conditions
		if isUnavailable || hasMotorIssue || hasBlockingObstruction {
			continue
		}

		stateMap[strconv.Itoa(id)] = HRAElevState{
			Behaviour:   elev.State.Behaviour.ToString(),
			Floor:       elev.State.Floor,
			Direction:   elev.State.Direction.ToString(),
			CabRequests: elev.CabRequests,
		}
	}
	if len(stateMap) == 0 {
		panic("No elevator states available for assignment!")
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

	return (*output)[strconv.Itoa(localID)]
}
