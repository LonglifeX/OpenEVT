package main

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/brandon1024/evt-client/internal/evt"
	"github.com/brandon1024/evt-client/internal/prom"
	"github.com/brandon1024/evt-client/internal/types"
)

func inverterConnect(ctx context.Context, client *evt.Client, reconnectInverval time.Duration) error {
	for {
		connect(ctx, client)

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
		if err != nil && !errors.Is(err, evt.ErrFrameDiscarded) {
			return err
		}

		// we received something we cant parse, skip over it
		if errors.Is(err, evt.ErrFrameDiscarded) {
			continue
		}

		log.Println("====================================================================================================")
		log.Writer().Write([]byte(msg.String()))
		log.Println("====================================================================================================")

		prom.Update(client.Address, &msg)
	}
}
