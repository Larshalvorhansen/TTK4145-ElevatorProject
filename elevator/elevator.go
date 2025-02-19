package elevator

import (
	"Driver-go/elevio"
	"fmt"
	"time"
)

// // Constants
const Num_floors = 4
const Door_open_duration = 3 * time.Second

// // Queue
type Queue_element struct {
	Floor  int
	Button elevio.ButtonType
}

type Queue struct {
	Orders   []Queue_element
	Capacity int
}

func Create_queue(capacity int) *Queue {
	return &Queue{
		Orders:   make([]Queue_element, 0, capacity),
		Capacity: capacity,
	}
}

func (q *Queue) Length() int {
	return len(q.Orders)
}

func (q *Queue) Add(order Queue_element) {
	//Check for duplicates
	exsist := false
	for _, ord := range q.Orders {
		if ord == order {
			exsist = true
		}
	}

	if q.Length() <= q.Capacity && !exsist {
		q.Orders = append(q.Orders, order)
	}
}

func (q *Queue) Remove(floor int) {
	temp := make([]Queue_element, 0, q.Capacity)
	for _, order := range q.Orders {
		if order.Floor != floor {
			temp = append(temp, order)
		}
	}
	q.Orders = temp
}

func (q *Queue) Empty() {
	q.Orders = q.Orders[:0]
}

func (q *Queue) Print() {
	fmt.Printf("Queue: \n")
	for i, order := range q.Orders {
		switch order.Button {
		case elevio.BT_HallUp:
			fmt.Printf("Order %d: floor %d, hall up button. \n", i, order.Floor)
		case elevio.BT_HallDown:
			fmt.Printf("Order %d: floor %d, hall down button. \n", i, order.Floor)
		case elevio.BT_Cab:
			fmt.Printf("Order %d: floor %d, cab button. \n", i, order.Floor)
		}
	}
}

func (q *Queue) Find_lowest_between(lower_bound int) int {
	lowest := q.Orders[0].Floor
	for _, order := range q.Orders[1:] {
		if (order.Floor < lowest && order.Floor >= lower_bound) && (order.Button == elevio.BT_HallUp || order.Button == elevio.BT_Cab) {
			lowest = order.Floor
		}
	}
	return lowest
}

func (q *Queue) Find_highest_between(upper_bound int) int {
	highest := q.Orders[0].Floor
	for _, order := range q.Orders[1:] {
		if (order.Floor > highest && order.Floor <= upper_bound) && (order.Button == elevio.BT_HallDown || order.Button == elevio.BT_Cab) {
			highest = order.Floor
		}
	}
	return highest
}

// // Elevator
type State int

const (
	Undefined State = iota
	Idle
	Moving
	FloorStop
	FullStop
)

type Elevator struct {
	Current_state State
	Current_floor int
	Direction     elevio.MotorDirection
	Obstruction   bool
	Timer_done    bool
	Queue         *Queue
	ID            int
}

func CreateElevator(id int, capacity int) *Elevator {
	return &Elevator{
		Current_state: Undefined,
		Queue:         Create_queue(capacity),
		ID:            id,
	}
}

func (e *Elevator) Add_order(order Queue_element) {
	e.Queue.Add(order)
	elevio.SetButtonLamp(order.Button, order.Floor, true)
	e.Queue.Print()
}

func (e *Elevator) Remove_order(order_floor int) {
	e.Queue.Remove(order_floor)
	e.Queue.Print()
}

func Reset_lamps_at_floor(floor int) {
	elevio.SetButtonLamp(elevio.BT_Cab, floor, false)
	elevio.SetButtonLamp(elevio.BT_HallUp, floor, false)
	elevio.SetButtonLamp(elevio.BT_HallDown, floor, false)
}

func Reset_all_lamps(floor_num int) {
	for floor := 0; floor < floor_num; floor++ {
		Reset_lamps_at_floor(floor)
	}
}

func (e *Elevator) Starting_routine(floor_sensor chan int) {
	if e.Current_floor < Num_floors-1 {
		elevio.SetMotorDirection(elevio.MD_Down)
		e.Current_floor = <-floor_sensor

		elevio.SetFloorIndicator(e.Current_floor)
		elevio.SetMotorDirection(elevio.MD_Stop)
	}
	e.Direction = elevio.MD_Stop
	e.Current_state = Idle
}

func (e *Elevator) MoveToFloor(target_floor int, doorTimer *time.Timer) {
	doorTimer.Reset(Door_open_duration)
	if e.Current_floor == target_floor && elevio.GetFloor() != -1 {
		elevio.SetMotorDirection(elevio.MD_Stop)
		e.Direction = elevio.MD_Stop

		e.Current_floor = elevio.GetFloor()

		e.Current_state = FloorStop
	} else if target_floor > e.Current_floor {
		elevio.SetMotorDirection(elevio.MD_Up)
		e.Direction = elevio.MD_Up
	} else if target_floor < e.Current_floor {
		elevio.SetMotorDirection(elevio.MD_Down)
		e.Direction = elevio.MD_Down
	}
}

func (e *Elevator) OpenDoor(doorTimer *time.Timer) {
	elevio.SetDoorOpenLamp(true)
	if e.Obstruction {
		doorTimer.Reset(Door_open_duration)
	}

	if !e.Obstruction && e.Timer_done {
		e.Timer_done = false
		elevio.SetDoorOpenLamp(false)
		if e.Queue.Length() > 0 {
			e.Current_state = Moving
		} else {
			e.Current_state = Idle
		}
	}
}

func (e *Elevator) StateMachine(floor_sensor chan int, numFloors int, doorTimer *time.Timer) {
	switch e.Current_state {
	case Undefined:
		e.Starting_routine(floor_sensor)

	case Idle:
		doorTimer.Reset(Door_open_duration)
		if e.Queue.Length() > 0 {
			e.Current_state = Moving
		}

	case Moving:
		doorTimer.Reset(Door_open_duration)
		stopFloor := e.Queue.Orders[0].Floor

		if stopFloor > e.Current_floor {
			stopFloor = e.Queue.Find_lowest_between(e.Current_floor)
		} else if stopFloor < e.Current_floor {
			stopFloor = e.Queue.Find_highest_between(e.Current_floor)
		}
		e.MoveToFloor(stopFloor, doorTimer)

	case FloorStop:
		if e.Direction == elevio.MD_Stop {
			e.Queue.Remove(e.Current_floor)
			Reset_lamps_at_floor(e.Current_floor)
			e.OpenDoor(doorTimer)
		}

	case FullStop:
		doorTimer.Reset(Door_open_duration)
		elevio.SetMotorDirection(elevio.MD_Stop)
		e.Direction = elevio.MD_Stop
		time.Sleep(2 * time.Second)
		e.Current_state = Undefined
		doorTimer.Reset(Door_open_duration)
	}
}

func RunSingleElevator(id int) {
	elevator := CreateElevator(id, Num_floors*3)
	elevio.Init("localhost:15657", Num_floors)
	Reset_all_lamps(Num_floors)
	elevio.SetDoorOpenLamp(false)

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstruction := make(chan bool)
	drv_stop := make(chan bool)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstruction)
	go elevio.PollStopButton(drv_stop)

	doorTimer := time.NewTimer(Door_open_duration)
	doorTimer.Stop()

	fmt.Printf("Elevator %d ready for use \n", elevator.ID)

	for {
		select {
		case floor := <-drv_floors:
			elevator.Current_floor = floor
			elevio.SetFloorIndicator(floor)

		case btn := <-drv_buttons:
			elevator.Add_order(Queue_element{btn.Floor, btn.Button})

		case <-drv_stop:
			elevator.Current_state = FullStop

		case obs := <-drv_obstruction:
			elevator.Obstruction = obs

		case <-doorTimer.C:
			elevator.Timer_done = true

		default:
			elevator.StateMachine(drv_floors, Num_floors, doorTimer)
		}
	}
}
