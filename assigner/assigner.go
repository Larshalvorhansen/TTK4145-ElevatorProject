package assigner

import (
	"Driver-go/config"
	"Driver-go/distributor"
	"Driver-go/elevator"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
)

type HRAState struct {
	Behaviour   string                 `json:"behaviour"`
	Floor       int                    `json:"floor"`
	Direction   string                 `json:"direction"`
	CabRequests [config.NumFloors]bool `json:"cabRequests"`
}

type HRAInput struct {
	HallRequests [config.NumFloors][2]bool `json:"hallRequests"`
	States       map[string]HRAState       `json:"states"`
}

func CalculateOptimalOrders(cs distributor.CommonState, id int) elevator.Orders {

	stateMap := make(map[string]HRAState)
	for i, v := range cs.States {
		if cs.Ackmap[i] == distributor.NotAvailable || v.State.Motorstop { // removed the additional "... || v.State,Obstructed" for single elevator use
			continue
		} else {
			stateMap[strconv.Itoa(i)] = HRAState{
				Behaviour:   v.State.Behaviour.ToString(),
				Floor:       v.State.Floor,
				Direction:   v.State.Direction.ToString(),
				CabRequests: v.CabRequests,
			}
		}
	}

	// For debugging
	if len(stateMap) == 0 {
		fmt.Println("no elevator states available for assignment!")
		panic("no elevator states available for assignment!")
	}

	hraInput := HRAInput{cs.HallRequests, stateMap}

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
