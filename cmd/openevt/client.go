package main

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"github.com/brandon1024/evt-client/internal/evt"
	"github.com/brandon1024/evt-client/internal/prom"
	"github.com/brandon1024/evt-client/internal/types"
)

func inverterConnect(ctx context.Context, client *evt.Client, reconnectInverval time.Duration) error {
	for {
		err := connect(ctx, client)

		log.Printf("INFO - connection lost to inverter %s; retrying in %s", client.InverterID, reconnectInverval.String())

		tm := time.NewTimer(reconnectInverval)

		select {
		case <-ctx.Done():
			tm.Stop()
			return ctx.Err()
		case <-tm.C:
			break
		}
	}
}

func connect(ctx context.Context, client *evt.Client) error {
	log.Printf("INFO - opening tcp connection to inverter %s at %s", client.InverterID, client.Address)

	// Connect to the inverter
	err := client.Connect()
	if err != nil {
		return err
	}

	log.Printf("INFO - connection established [%s]", client.String())

	defer client.Close()

	prom.UpdateConnectionStatus(client.Address, client.InverterID, 1.0)
	defer prom.UpdateConnectionStatus(client.Address, client.InverterID, 0.0)

	go func() {
		<-ctx.Done()
		client.Close()
	}()

	// first, poll for current state
	err = client.Poll()
	if err != nil {
		return err
	}

	// setup read loop
	for {
		var msg types.InverterStatus

		err = client.ReadFrame(&msg)

		// if we reached the deadline, poll
		if errors.Is(err, os.ErrDeadlineExceeded) {
			err = client.Poll()
			if err != nil {
				return err
			}

			continue
		}

		// we received something we cant parse, skip over it
		if errors.Is(err, evt.ErrFrameDiscarded) {
			continue
		}

		if err != nil {
			return err
		}

		log.Printf("DEBUG - inverter status message received [%fW %fkWh]",
			msg.Module1.OutputPowerAC+msg.Module2.OutputPowerAC,
			msg.Module1.TotalEnergy+msg.Module2.TotalEnergy,
		)

		prom.Update(client.Address, &msg)
	}
}
