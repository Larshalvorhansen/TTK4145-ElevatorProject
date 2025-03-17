package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"
)

const stateFile = "/tmp/counter_state.txt" // File to store last count

// Reads last saved count from the file
func readLastCount() int {
	data, err := os.ReadFile(stateFile)
	if err != nil {
		return 1 // Start from 1 if file doesn't exist
	}
	count, err := strconv.Atoi(string(data))
	if err != nil {
		return 1
	}
	return count
}

// Saves current count to the file
func saveCount(count int) {
	os.WriteFile(stateFile, []byte(strconv.Itoa(count)), 0644)
}

// Function to start the counter process
func startCounter() {
	count := readLastCount()

	// Goroutine to save count periodically
	go func() {
		for {
			time.Sleep(1 * time.Second)
			saveCount(count)
		}
	}()

	// Infinite loop to count from 1 to 4
	for {
		fmt.Println(count)
		count++
		if count > 4 {
			count = 1
		}
		time.Sleep(1 * time.Second)
	}
}

// Opens a new terminal and restarts the program in it
func restartInNewTerminal() {
	// Detect OS and use appropriate terminal command
	var cmd *exec.Cmd
	if _, err := exec.LookPath("gnome-terminal"); err == nil {
		// Linux (GNOME-based)
		cmd = exec.Command("gnome-terminal", "--", os.Args[0], "child")
	} else if _, err := exec.LookPath("xfce4-terminal"); err == nil {
		// XFCE Terminal
		cmd = exec.Command("xfce4-terminal", "--hold", "--command="+os.Args[0]+" child")
	} else if _, err := exec.LookPath("konsole"); err == nil {
		// KDE Konsole
		cmd = exec.Command("konsole", "-e", os.Args[0], "child")
	} else if _, err := exec.LookPath("x-terminal-emulator"); err == nil {
		// Generic Linux terminal
		cmd = exec.Command("x-terminal-emulator", "-e", os.Args[0], "child")
	} else if _, err := exec.LookPath("osascript"); err == nil {
		// macOS Terminal
		cmd = exec.Command("osascript", "-e", `tell application "Terminal" to do script "`+os.Args[0]+` child"`)
	} else if _, err := exec.LookPath("cmd.exe"); err == nil {
		// Windows Command Prompt
		cmd = exec.Command("cmd.exe", "/C", "start", os.Args[0], "child")
	} else {
		fmt.Println("No compatible terminal found. Restart manually.")
		return
	}

	cmd.Start()
}

func main() {
	// If child process, start counter
	if len(os.Args) > 1 && os.Args[1] == "child" {
		startCounter()
		return
	}

	// Parent process: Monitors child and restarts it in a new terminal if it stops
	for {
		// Start child process
		cmd := exec.Command(os.Args[0], "child")
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true} // Separate process group
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err := cmd.Start()
		if err != nil {
			fmt.Println("Failed to start child process:", err)
			time.Sleep(2 * time.Second)
			continue
		}

		// Wait for child to exit
		err = cmd.Wait()
		fmt.Println("Child process stopped. Restarting in a new terminal...")

		// Open a new terminal with the program running
		restartInNewTerminal()

		time.Sleep(2 * time.Second) // Small delay before restart
	}
}
