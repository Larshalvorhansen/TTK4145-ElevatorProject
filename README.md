# Elevator Project – Group 15

This is a distributed elevator control system developed as part of the project in TTK4145 Real-Time Programming.

The project description can be found [here](https://github.com/TTK4145/Project).

## Table of Contents

- [Overview](#overview)
- [How to run](#how-to-run)
- [Module Structure](#module-structure)
- [Credits](#credits)

## Overview

This system is designed to coordinate multiple elevators in a networked environment. Its core responsibilities include:

- Distributing hall requests efficiently using a cost-based assignment strategy.
- Keeping all elevator processes synchronized through peer-to-peer UDP broadcasting.
- Detecting disconnected or crashed peers and adjusting state accordingly.
- Allowing elevators to continue operating locally in the absence of network communication.

Each elevator node maintains a shared view of the system state, enabling coordinated behavior without relying on a central server.

## How to Run
Before running the program, ensure that the parameters in `config/config.go` are set to the required values. Update them if necessary.

### Linux (Recommended)

Use the provided `start.sh` script to build and start the elevator program interactively:
1. Make the script executable (only needed once):
    ```bash
    chmod +x start.sh
    ```
2. Start the program:
    ```bash
    ./start.sh
    ```

### Manual Start / Other Operating Systems
Running the program:
```bash
go run main.go
```
Runs the program with default values on ID (0) and port (15657)

For changing ID and Port run:
```bash
go run main.go -id=0 -port=15657
```

## Module Structure

This system consists of several modules, each responsible for specific functionality:

### `start.sh`
Shell script for building and launching the elevator program.  
Prompts for elevator ID and port, and restarts the program automatically if it crashes.

### `main.go`
Entry point of the program. Initializes all channels and modules, and connects the main control loop.

### `assigner/`
Implements the logic for assigning hall orders to elevators based on a cost function.  
Determines which elevator is best suited to handle each request.

### `config/`
Defines system-wide constants such as the number of floors, number of elevators, channel buffer sizes, and timing configurations.

### `coordinator/`
Maintains and synchronizes the shared system state across all active elevators.  
Implements the core state machine for coordinating elevator actions and handling peer communication.

### `elevator/`
Handles local elevator behavior and state.  
Controls door logic, motor direction, floor tracking, and manages the elevator’s internal state machine.

### `hardware/`
Provides an interface to the elevator simulator.  
Includes functions for motor control, button polling, floor sensors, and lamp updates.

### `lamp/`
Defines functionality for controlling the button lamps.
Updates are based on the current shared system state.

### `network/`
Handles UDP-based peer-to-peer communication:
- `bcast/`: Encodes and broadcasts typed JSON messages
- `conn/`: Initializes and manages UDP sockets across platforms
- `localip/`: Retrieves the IP address of the current machine
- `peers/`: Tracks active peers, handles peer discovery and loss detection

## Credits

This project contains code based on example code and resources from the [TTK4145 Real-Time Programming course](https://github.com/TTK4145).  
Files that include adapted example code, or from ChatGPT (OpenAI), are marked either with a README or with comments at the top, indicating the source of the original code and describing any modifications made.