package main

import (
    "fmt"
    "os"
    "os/exec"
    "os/signal"
    "syscall"
)

func main() {
    // Determine which mode we’re running in: “main” or “fallback”
    if len(os.Args) > 1 && os.Args[1] == "fallback" {
        fmt.Println("Running Fallback. Press Ctrl+C to spawn Main in a new terminal.")
        waitForInterrupt("main")
    } else {
        fmt.Println("Running Main. Press Ctrl+C to spawn Fallback in a new terminal.")
        waitForInterrupt("fallback")
    }
}

func waitForInterrupt(nextMode string) {
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

    // Block until we receive an interrupt signal (Ctrl+C)
    <-sigChan
    fmt.Printf("\nInterrupt received. Spawning %s in a new terminal...\n", nextMode)

    // Command to launch a new terminal and run this program in the opposite mode
    cmd := exec.Command("gnome-terminal", "--", "go", "run", "main.go", nextMode)
    if err := cmd.Start(); err != nil {
        fmt.Println("Error launching new terminal:", err)
    }

    // Exit current process so only the new one keeps running
    os.Exit(0)
}
