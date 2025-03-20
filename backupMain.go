package main

import (
	"Driver-go/assigner"
	"Driver-go/config"
	"Driver-go/distributor"
	"Driver-go/elevator"
	"Driver-go/hardware"
	"Driver-go/lamp"
	"Driver-go/network/bcast"
	"Driver-go/network/peers"
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"
	"time"
)

const filename = "state.csv"
const restartProgram = "main.go"

func openTerminalAndRun(command string, port, id int) {
	var cmd *exec.Cmd
	args := fmt.Sprintf("go run %s -port %d -id %d", command, port, id)

	if runtime.GOOS == "darwin" {
		cmd = exec.Command("osascript", "-e", fmt.Sprintf("tell application \"Terminal\" to do script \"%s\"", args))
	} else {
		cmd = exec.Command("x-terminal-emulator", "-e", "bash", "-c", fmt.Sprintf("%s; exec bash", args))
	}

	cmd.Start()
	os.Exit(0)
}

func readLastState() int {
	file, err := os.Open(filename)
	if err != nil {
		return 1 // Start from 1 if file doesn't exist
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil || len(records) == 0 {
		return 1
	}

	lastVal, err := strconv.Atoi(records[len(records)-1][0])
	if err != nil {
		return 1
	}

	return lastVal
}

func writeStateToCSV(value int) {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{strconv.Itoa(value)})
}

func main() {
	serverPort := flag.Int("port", 15657, "Elevator server port (default: 15657)")
	elevatorId := flag.Int("id", 0, "Elevator ID (default: 0)")
	flag.Parse()

	port := *serverPort
	id := *elevatorId

	hardware.Init("localhost:"+strconv.Itoa(port), config.NumFloors)

	fmt.Printf("Elevator system started successfully!\n  Elevator Details:\n\tID:   %d\n\tPort: %d\n  System Configuration:\n\tFloors:    %d\n\tElevators: %d\n\n", id, port, config.NumFloors, config.NumElevators)

	newOrderC := make(chan elevator.Orders, config.BufferSize)
	deliveredOrderC := make(chan hardware.ButtonEvent, config.BufferSize)
	newStateC := make(chan elevator.State, config.BufferSize)
	confirmedCommonStateC := make(chan distributor.CommonState, config.BufferSize)
	networkTxC := make(chan distributor.CommonState, config.BufferSize)
	networkRxC := make(chan distributor.CommonState, config.BufferSize)
	peersRxC := make(chan peers.PeerUpdate, config.BufferSize)
	peersTxC := make(chan bool, config.BufferSize)

	go peers.Receiver(config.PeersPortNumber, peersRxC)
	go peers.Transmitter(config.PeersPortNumber, id, peersTxC)
	go bcast.Receiver(config.BcastPortNumber, networkRxC)
	go bcast.Transmitter(config.BcastPortNumber, networkTxC)
	go distributor.Distributor(confirmedCommonStateC, deliveredOrderC, newStateC, networkTxC, networkRxC, peersRxC, id)
	go elevator.Elevator(newOrderC, deliveredOrderC, newStateC)

	state := readLastState()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Goroutine to track state
	go func() {
		for {
			fmt.Println("Recorded state:", state)
			writeStateToCSV(state)
			state++
			if state > 4 {
				state = 1
			}
			time.Sleep(1 * time.Second)
		}
	}()

	// Goroutine to handle interrupts
	go func() {
		for sig := range c {
			fmt.Println("Received signal:", sig)
			fmt.Println("Interrupted! Restarting", restartProgram, "in new terminal...")
			openTerminalAndRun(restartProgram, port, id)
		}
	}()

	for {
		select {
		case commonState := <-confirmedCommonStateC:
			newOrderC <- assigner.CalculateOptimalOrders(commonState, id)
			lamp.SetLamps(commonState, id)
		default:
			continue
		}
	}
}
