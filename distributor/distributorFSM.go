package distributor

import (
	"Driver-go/config"
	"time"
)

/*
Logisk oversikt:


->sjekk disconect timer
->sende hearbeat
->sende status
-> ta en oppgave hvis du kan

ikke koblet til
	ferdig med alt
		-> sjekk om reconectet
			break
	optatt med ordre
		-> Gjør ferdig ordre

koblet til
	-> reset disconect timer
	idle
		-> lytte etter ordre
	ikke idle
		-> gjør ferdig ordre

*/

func Distributor(


){
	disconnectTimer := time.NewTimer(config.DisconnectTime)
	updateInterval := time.NewTicker(config.UpdateIntervalTime)
	
	idle := true
	online := true

	for {
		select {
		//Sjekke disconnect timer
		case :/*noe skrives på disconnect timer kanalen?*/
			//setter alle andre heiser en "denne heisen" som unavalible
			fmt.Println("Status offline")
			online = false

		//Send status
		case :/*noe skrives på tickerkanalen*/
			sende common state til networkTx
			
		case :/*noe må gjøres*/:

			/*
			->sende status
			-> ta en oppgave hvis du kan*/

		default:
		}

		switch {
		case !online:
			//Styre med våre egne heis cabcalls
			select {
			case temp://mottar noe fra networkRX
				online = true
				break
			}

			// Motta nye ordre

			// Ferdig med ordre

			// Ny state
		}
	}

}
