package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"time"

	"github.com/brandon1024/OpenEVT/internal/evt"
	"github.com/brandon1024/OpenEVT/internal/types"
	"github.com/brandon1024/OpenEVT/internal/web"
)

func inverterConnect(ctx context.Context, client *evt.Client, reconnectInverval time.Duration) error {
	for {
		connect(ctx, client)

		slog.Info("connection lost to inverter; retrying...",
			"serial", client.InverterID,
			"retry-interval", reconnectInverval.String(),
		)

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
	slog.Info("opening tcp connection to inverter", "serial", client.InverterID, "address", client.Address)

	// Connect to the inverter
	err := client.Connect()
	if err != nil {
		return err
	}

	slog.Info("connection established", "status", client.String())

	defer client.Close()

	web.UpdateConnectionStatus(client.Address, client.InverterID, 1.0)
	defer web.UpdateConnectionStatus(client.Address, client.InverterID, 0.0)

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

		slog.Debug("inverter status message received",
			"power-ac", msg.Module1.OutputPowerAC+msg.Module2.OutputPowerAC,
			"total-energy", msg.Module1.TotalEnergy+msg.Module2.TotalEnergy,
		)

		web.Update(client.Address, &msg)
	}
}
