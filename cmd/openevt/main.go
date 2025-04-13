package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/brandon1024/evt-client/internal/evt"
	"github.com/brandon1024/evt-client/internal/types"
)

func main() {
	client := &evt.Client{}

	flag.StringVar(&client.InverterID, "serial-number", "", "serial number of your microinverter (e.g. 31583078)")
	flag.StringVar(&client.Address, "addr", "", "address and port of the microinverter (e.g. 192.0.2.1:14889)")
	flag.Parse()

	log.Printf("INFO - opening tcp connection to inverter %s at %s", client.InverterID, client.Address)

	// Connect to the inverter
	err := client.Connect()
	if err != nil {
		fmt.Printf("ERROR - failed to connect [%v]\n", err)
		os.Exit(1)
	}

	log.Printf("INFO - connection established [%s]", client.String())

	defer client.Close()

	// first, poll for current state
	err = client.Poll()
	if err != nil {
		fmt.Printf("ERROR - failed to poll inverter [%v]\n", err)
		os.Exit(1)
	}

	// setup read loop
	for {
		var msg types.InverterStatus

		err = client.ReadFrame(&msg)
		if err != nil && !errors.Is(err, evt.ErrFrameDiscarded) {
			fmt.Printf("ERROR - failed to read frame from inverter [%v]\n", err)
			os.Exit(1)
		}

		// we received something we cant parse, just poll to keep the connection active
		if errors.Is(err, evt.ErrFrameDiscarded) {
			err = client.Poll()
			if err != nil {
				fmt.Printf("ERROR - failed to poll inverter [%v]\n", err)
				os.Exit(1)
			}

			continue
		}

		log.Println("====================================================================================================")
		log.Writer().Write([]byte(msg.String()))
		log.Println("====================================================================================================")
	}
}
