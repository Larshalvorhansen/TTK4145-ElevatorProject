package main

import (
	"fmt"
	"sort"
	"time"
)

const tick = time.Millisecond * 33

// Ressurstypen – en liste over hvilke ID-er som har brukt ressursen
type Resource struct {
	value []int
}

// En forespørsel om ressurs fra en bruker
type ResourceRequest struct {
	id       int
	priority int
	channel  chan Resource
}

// Ressurshåndterer: sender ressursen til de med høyest prioritet først
func resourceManager(askFor chan ResourceRequest, giveBack chan Resource) {
	res := Resource{}
	busy := false
	queue := PriorityQueue{}

	for {
		select {
		// Tar imot forespørsel og legger i kø etter prioritet
		case request := <-askFor:
			queue.Insert(request, request.priority)

		// Tar imot ressursen tilbake fra en bruker
		case res = <-giveBack:
			busy = false
		}

		// Hvis ressursen er ledig og det finnes forespørsler, send til neste i køen
		if !busy && !queue.Empty() {
			request := queue.Front().(ResourceRequest)
			queue.PopFront()
			request.channel <- res
			busy = true
		}
	}
}

// Konfigurasjon for én ressursbruker
type ResourceUserConfig struct {
	id        int // ID for bruker
	priority  int // 0 = lav, 1 = høy
	release   int // når brukeren skal be om ressurs
	execution int // hvor lenge brukeren holder ressursen
}

// Rutine for hver ressursbruker
func resourceUser(cfg ResourceUserConfig, askFor chan ResourceRequest, giveBack chan Resource) {
	replyChan := make(chan Resource)

	time.Sleep(time.Duration(cfg.release) * tick)

	executionStates[cfg.id] = waiting
	askFor <- ResourceRequest{cfg.id, cfg.priority, replyChan}
	res := <-replyChan

	executionStates[cfg.id] = executing

	time.Sleep(time.Duration(cfg.execution) * tick)
	res.value = append(res.value, cfg.id)
	giveBack <- res

	executionStates[cfg.id] = done
}

func main() {
	askFor := make(chan ResourceRequest, 10)
	giveBack := make(chan Resource)
	go resourceManager(askFor, giveBack)

	executionStates = make([]ExecutionState, 10)

	cfgs := []ResourceUserConfig{
		{0, 0, 1, 1},
		{1, 0, 3, 1},
		{2, 1, 5, 1},

		{0, 1, 10, 2},
		{1, 0, 11, 1},
		{2, 1, 11, 1},
		{3, 0, 11, 1},
		{4, 1, 11, 1},
		{5, 0, 11, 1},
		{6, 1, 11, 1},
		{7, 0, 11, 1},
		{8, 1, 11, 1},

		{0, 1, 25, 3},
		{6, 0, 26, 2},
		{7, 0, 26, 2},
		{1, 1, 26, 2},
		{2, 1, 27, 2},
		{3, 1, 28, 2},
		{4, 1, 29, 2},
		{5, 1, 30, 2},
	}

	go executionLogger()

	// Start alle brukerne parallelt
	for _, cfg := range cfgs {
		go resourceUser(cfg, askFor, giveBack)
	}

	// Simuleringen avsluttes med å hente ressursen og vise rekkefølgen
	time.Sleep(time.Duration(45) * tick)

	resourceCh := make(chan Resource)
	askFor <- ResourceRequest{0, 1, resourceCh}
	executionOrder := <-resourceCh
	fmt.Println("Execution order:", executionOrder)
}

// --- PRIORITETSKØ --- //

// En kø som holder elementer i prioritert rekkefølge.
// Høyere tall = høyere prioritet. FIFO for lik prioritet.
type PriorityQueue struct {
	queue []struct {
		val      interface{}
		priority int
	}
}

func (pq *PriorityQueue) Insert(value interface{}, priority int) {
	pq.queue = append(pq.queue, struct {
		val      interface{}
		priority int
	}{value, priority})

	sort.SliceStable(pq.queue, func(i, j int) bool {
		return pq.queue[i].priority > pq.queue[j].priority
	})
}

func (pq *PriorityQueue) Front() interface{} {
	return pq.queue[0].val
}

func (pq *PriorityQueue) PopFront() {
	pq.queue = pq.queue[1:]
}

func (pq *PriorityQueue) Empty() bool {
	return len(pq.queue) == 0
}

// --- VISNING AV BRUKERSTATUS --- //

type ExecutionState rune

const (
	none      ExecutionState = '\u0020'
	waiting                  = '\u2592'
	executing                = '\u2593'
	done                     = '\u2580'
)

var executionStates []ExecutionState

// Skriver ut status for alle brukere hvert tick
func executionLogger() {
	time.Sleep(tick / 2)
	t := 0

	fmt.Printf("  id:")
	for id := range executionStates {
		fmt.Printf("%3d", id)
	}
	fmt.Printf("\n")

	for {
		grid := ' '
		if t%5 == 0 {
			grid = '\u2500'
		}

		fmt.Printf("%04d : ", t)
		for id, state := range executionStates {
			fmt.Printf("%c%c%c", state, grid, grid)
			if state == done {
				executionStates[id] = none
			}
		}
		fmt.Printf("\n")
		t++
		time.Sleep(tick)
	}
}
