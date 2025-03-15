# TTK4145-ElevatorProject

This is the repo for the elevator project, group 15.

This project implements an elevator control system in Go, consisting of a peer to peer network of multiple modules handling different functionalities of the elevator. 

Below follows a description of each module:

## Modules
### **elevator**
Handles the core functionality of the elevator, including movement, door control, order handling, and finite state machine (FSM) logic.

### **elevio**
Provides an interface for hardware interaction.

### **lights**
Manages the incicator lights for the elevator system.

### **assigner**
TODO

### **network**
TTK4145 developed this network module. It can be accessed [here](https://github.com/TTK4145/Network-go). We have adjusted the code slightly to suit our specific implementation. Still, it remains mostly unchanged from the original.

### **config**
Stores configuration parameters that other modules can use.

More to add?
