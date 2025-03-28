package main

import "fmt"
import "time"

const tick = time.Millisecond * 33

// En ressurs som inneholder en liste over IDene til de som har brukt den
type Resource struct {
	value []int
}

// Ressursbehandleren, denne skal sørge for at brukere får tilgang til ressursen
// Prioriterer høyt prioriterte forespørsler
func resourceManager(takeLow chan Resource, takeHigh chan Resource, giveBack chan Resource) {
	res := Resource{}

	for {
		// Først: vent på at ressursen skal komme tilbake før vi prøver å sende den ut igjen
		res = <-giveBack

		// Etter at ressursen er tilbake, send den til neste bruker
		for {
			select {
			// Forsøk å gi til høy prioritet først
			case takeHigh <- res:
				goto next // ressursen er sendt, gå videre til neste runde
			// Hvis ingen høy prioritet venter, prøv lav prioritet
			case takeLow <- res:
				goto next
			// Hvis fortsatt ingen er klare, bare vent litt før vi prøver igjen
			default:
				time.Sleep(tick / 10)
			}
		}
	next:
	}
}

// Konfigurasjon for én ressursbruker
type ResourceUserConfig struct {
	id        int // Brukerens ID
	priority  int // 1 = høy prioritet, 0 = lav
	release   int // Når brukeren skal prøve å hente ressursen
	execution int // Hvor lenge brukeren holder ressursen
}

// En rutine som simulerer en bruker som venter, bruker og returnerer ressursen
func resourceUser(cfg ResourceUserConfig, take chan Resource, giveBack chan Resource) {
	time.Sleep(time.Duration(cfg.release) * tick)

	executionStates[cfg.id] = waiting
	res := <-take

	executionStates[cfg.id] = executing
	time.Sleep(time.Duration(cfg.execution) * tick)

	res.value = append(res.value, cfg.id)
	giveBack <- res

	executionStates[cfg.id] = done
}

func main() {
	takeLow := make(chan Resource)
	takeHigh := make(chan Resource)
	giveBack := make(chan Resource)

	go resourceManager(takeLow, takeHigh, giveBack)

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

	// Start alle brukere parallelt
	for _, cfg := range cfgs {
		if cfg.priority == 1 {
			go resourceUser(cfg, takeHigh, giveBack)
		} else {
			go resourceUser(cfg, takeLow, giveBack)
		}
	}

	// Ingen ordentlig måte å vente på at alt er ferdig – vi bruker Sleep
	time.Sleep(time.Duration(45) * tick)

	executionOrder := <-takeHigh
	fmt.Println("Execution order:", executionOrder)
}

// Tilstandene en bruker kan være i under simulering
type ExecutionState rune

const (
	none      ExecutionState = '\u0020' // Ingenting
	waiting                  = '\u2592' // Venter på ressurs
	executing                = '\u2593' // Bruker ressurs
	done                     = '\u2580' // Ferdig
)

var executionStates []ExecutionState

// Logger tilstandene til alle brukerne i terminalen
func executionLogger() {
	time.Sleep(tick / 2)
	t := 0

	// Skriv ut IDene til brukerne
	fmt.Printf("  id:")
	for id := range executionStates {
		fmt.Printf("%3d", id)
	}
	fmt.Printf("\n")

	for {
		grid := ' '
		if t%5 == 0 {
			grid = '\u2500' // horisontalt rutenett
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
