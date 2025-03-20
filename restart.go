package main

import (
	"encoding/csv"
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
const restartProgram = "backupRestart.go"

func openTerminalAndRun(command string) {
	var cmd *exec.Cmd

	if runtime.GOOS == "darwin" {
		cmd = exec.Command("osascript", "-e", fmt.Sprintf("tell application \"Terminal\" to do script \"go run %s\"", command))
	} else {
		cmd = exec.Command("x-terminal-emulator", "-e", "bash", "-c", fmt.Sprintf("go run %s; exec bash", command))
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
	fmt.Println("Running ", restartProgram, "...")

	state := readLastState()
	ch := make(chan int)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		for {
			ch <- state
			writeStateToCSV(state)
			fmt.Println("Recorded:", state)
			state++
			if state > 4 {
				state = 1
			}
			time.Sleep(1 * time.Second)
		}
	}()

	go func() {
		for sig := range c {
			fmt.Println("Received signal:", sig)
			fmt.Println("Interrupted! Restarting", restartProgram, "in new terminal...")
			openTerminalAndRun(restartProgram)
		}
	}()

	select {} // Keep the program running
}
