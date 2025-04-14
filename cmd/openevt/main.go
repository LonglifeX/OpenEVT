package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/brandon1024/evt-client/internal/evt"
	"github.com/brandon1024/evt-client/internal/prom"
)

func main() {
	var (
		webListenAddress       string
		telemetryPath          string
		disableExporterMetrics bool
		reconnectInverval      time.Duration
	)

	client := &evt.Client{}

	flag.StringVar(&client.InverterID, "serial-number", "", "serial number of your microinverter (e.g. 31583078)")
	flag.StringVar(&client.Address, "addr", "", "address and port of the microinverter (e.g. 192.0.2.1:14889)")
	flag.DurationVar(&client.ReadTimeout, "poll-interval", time.Duration(0), "attempt to poll the inverter status more frequently than advertised")
	flag.DurationVar(&reconnectInverval, "reconnect-interval", time.Minute, "interval between connection attempts (e.g. 1m)")

	flag.StringVar(&webListenAddress, "web.listen-address", ":9090", "address on which to expose metrics")
	flag.StringVar(&telemetryPath, "web.telemetry-path", "/metrics", "path under which to expose metrics")
	flag.BoolVar(&disableExporterMetrics, "web.disable-exporter-metrics", false, "exclude metrics about the exporter itself (go_*)")

	flag.Parse()

	grp, ctx := errgroup.WithContext(context.Background())
	grp.Go(func() error {
		return inverterConnect(ctx, client, reconnectInverval)
	})
	grp.Go(func() error {
		return prom.ListenAndServe(ctx, webListenAddress, telemetryPath, disableExporterMetrics)
	})

	if err := grp.Wait(); err != nil {
		fmt.Printf("ERROR - %v\n", err)
		os.Exit(1)
	}
}
