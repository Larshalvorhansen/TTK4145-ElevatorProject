# TTK4145-ElevatorProject

This is the repo for the elevator project, group 15

The project aims to implement an elevator control system in Go using a **peer-to-peer** system with UDP broadcasting for communication, which should be able to control 3 elevators over 4 floors in a reasonable way according to the given specification. The system consists of several modules which handles different functionalities of the system, ranging from communication between elevators to assigning hallorders to different elevators or turing on/off lights. Below is a desciption of each module:

## Modules

### **elevator**
Handles the core functionality of the elevator, including movement, door control, order handling, lamp handling, and finite state machine (FSM) logic.

### **elevio**
TTK4145 developed this network module, which you can access [here](https://github.com/TTK4145/driver-go). It provides an interface for interacting with the elevator hardware.
Although we have made minor adjustments to the code to fit our project, it remains largely unchanged.


### **assigner**
Assigns the hall requests to different elevators based on an algorithm using the provided example code in project resources which is accesible [here](https://github.com/TTK4145/Project-resources/tree/master/cost_fns). Should have a main function that takes the elevatorstates and hallrequests as input and outputs which request are assigned to which elevators.

### **sharedstate**
Is far from finished, but should create a struct which keeps track of the state of the whole system (all elevators and orders), as well as making sure all elevators have a syncronized worldview. This also includes functions to manange the sharedstate-struct. A final-state-machine should be implented in order to use this sharedstate in a logical way, so that it is in fact syncronized.

### **network**
TTK4145 developed this network module, which you can access [here](https://github.com/TTK4145/Network-go). 
Although we have made some minor modifications, the module is mostly unchanged.

### **config**
Stores configuration parameters that several modules use, e.g. number of floors.

| Sjekkliste                                                                                       | ja/nei  |
| ------------------------------------------------------------------------------------------------ | ------- |
| Hall-knappen lyser når trykket på                                                                | ja      |
| Heis ankommer etasjen etter hall-knapp er trykket                                                | ja |
| Cab-knappen lyser når trykket på                                                                 | ja |
| Heis tar imot cab-kall og kjører til riktig etasje                                               | ja |
| Heis mister ikke noen kall (hall eller cab)                                                      | ja |
| Heis fortsetter å fungere ved nettverksbrudd                                                     | kanskje |
| Heis fortsetter å fungere ved strømbrudd                                                         | ja |
| Heis fullfører cab-kall etter strøm/nettverk kommer tilbake                                      | ja |
| Heis håndterer feil innen noen sekunder (ikke minutter)                                          | kankskje |
| Ved nettverksbrudd fortsetter heis å betjene eksisterende kall                                   | nei? |
| Heis tar fortsatt nye cab-kall ved nettverksbrudd                                                | usikker |
| Heis trenger ikke manuell restart etter strøm/nettverk går tilbake                               | nei |
| Hall-knapper på forskjellige arbeidsstasjoner viser samme lys under normale forhold              | ja |
| Minst én hall-knapp viser riktig lys ved pakketap                                                | ja |
| Cab-knappelysene er ikke delt mellom heiser                                                      | kanskje |
| Knappelys skrur seg på raskt etter trykk                                                         | kanskje |
| Knappelys skrur seg av når kallet er utført                                                      | kanskje |
| Døren åpner seg når heisen stopper på etasjen                                                    | kanskje |
| “Dør åpen”-lampen er tent når døren er åpen                                                      | kanskje |
| Døren lukker seg ikke mens heisen beveger seg                                                    | kanskje |
| Døren holder seg åpen i 3 sekunder på etasjen                                                    | kanskje |
| Døren lukker seg ikke hvis en hindring er til stede                                              | kanskje |
| Heisen stopper ikke på hver etasje unødvendig                                                    | ja |
| Hall-knappelyset slukker når heis ankommer riktig retning                                        | kanskje |
| Heisen skifter ikke retning unødvendig                                                           | kanskje |
| Heisen annonserer retning korrekt (opp/ned)                                                      | kanskje |
| Hvis heisens retning endres, fjernes motsatt retningskall og døren holdes åpen 3 sekunder ekstra | kanskje |

minitest