package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
)

func openTerminalAndRun(command string) {
	var cmd *exec.Cmd

	if runtime.GOOS == "darwin" {
		// Open a new terminal window and run the program
		cmd = exec.Command("osascript", "-e", fmt.Sprintf(`tell application "Terminal" to do script "go run %s"`, command))
	} else {
		// Linux: Using gnome-terminal or x-terminal-emulator
		cmd = exec.Command("x-terminal-emulator", "-e", "bash", "-c", fmt.Sprintf("go run %s; exec bash", command))
	}

	// Start the new terminal window
	cmd.Start()

	// Exit the current process to prevent multiple overlapping processes
	os.Exit(0)
}

func main() {
	fmt.Println("Running restart.go...")

	// Handle SIGINT (Ctrl+C)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	fmt.Println("Interrupted! Restarting restart.go in new terminal...")
	openTerminalAndRun("restart.go")
}
