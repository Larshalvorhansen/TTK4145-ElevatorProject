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
const restartProgram = "restart.go"

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
		fmt.Println("No previous state found. Starting from 1.")
		return 1 // Start from 1 if file doesn't exist
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil || len(records) == 0 {
		fmt.Println("CSV file empty or unreadable. Starting from 1.")
		return 1
	}

	lastVal, err := strconv.Atoi(records[len(records)-1][0])
	if err != nil {
		fmt.Println("Invalid state in CSV. Resetting to 1.")
		return 1
	}

	fmt.Println("Resuming from last recorded state:", lastVal)
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

	err = writer.Write([]string{strconv.Itoa(value)})
	if err != nil {
		fmt.Println("Error writing to CSV:", err)
	}
}

func main() {
	fmt.Println("Running backupRestart.go...")

	state := readLastState()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Goroutine to update state and write to CSV every second
	go func() {
		for {
			fmt.Println("Recorded:", state) // Debugging output
			writeStateToCSV(state)
			state++
			if state > 4 {
				state = 1
			}
			time.Sleep(1 * time.Second)
		}
	}()

	// Goroutine to handle Ctrl+C (interrupt) signals
	go func() {
		for sig := range c {
			fmt.Println("Received signal:", sig)
			fmt.Println("Interrupted! Restarting", restartProgram, "in new terminal...")
			openTerminalAndRun(restartProgram)
		}
	}()

	select {} // Keep the program running indefinitely
}
