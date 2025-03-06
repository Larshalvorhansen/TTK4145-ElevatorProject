# TTK4145-ElevatorProject

This is the repo for the elevator project, group 15

This project implements an elevator control system in Go, consisting of multiple modules handling different functionalities of the elevator. Below is a description of each module:

## Modules

### **elevator**
Handles the core functionality of the elevator, including movement, door control, order handling, and finite state machine (FSM) logic.

### **elevio**
Provides an interface for hardware interaction.

### **lights**
Manages the lights for the elevator system.

### **assigner**
Assigns the hall requests to different elevators based on an algorithm using the provided example code in project resources which is accesible [here](https://github.com/TTK4145/Project-resources/tree/master/cost_fns).

### **network**
TTK4145 developed this network module, which you can access [here](https://github.com/TTK4145/Network-go). Although we adjusted the code slightly to suit our purposes, but it remains mostly unchanged from the original.

### **config**
Stores configuration parameters that several modules use, e.g. number of floors.


More to add?
