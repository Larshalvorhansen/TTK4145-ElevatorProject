# TTK4145-ElevatorProject

This is the repo for the elevator project, group 15

This project implements an elevator control system in Go, consisting of multiple modules handling different functionalities of the elevator. Below is a description of each module:

## Modules

### **elevator/**
Handles the core functionality of the elevator, including movement, door control, order handling, and finite state machine (FSM) logic.

- **direction.go**: Defines the `Direction` type (Up, Down) and provides utility functions for converting between directions and motor/button types.
- **door.go**: Manages the state of the elevator door, including opening, closing, and handling obstructions.
- **elevatorFSM.go**: Implements the elevator's finite state machine, managing state transitions based on orders, floor arrivals, and obstructions.
- **orders.go**: Defines the `Orders` type to track and process elevator orders, including handling completed orders.

### **elevio/**
Provides an interface for hardware interaction.

- **elevator_io.go**: Handles communication with the elevator hardware, including motor control, button polling, floor sensors, and status indicators.

### **lights/**
Manages the indicator lights for the elevator system.

- **lights.go**: Controls the lighting of hall and cab buttons based on the elevator's state.

### **config/**
Stores configuration parameters that other modules can use.
