package prom

import (
	"context"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func ListenAndServe(ctx context.Context, addr, path string, disableExporterMetrics bool) error {
	if !disableExporterMetrics {
		reg.MustRegister(collectors.NewGoCollector())
	}

	http.Handle(path, promhttp.HandlerFor(reg, promhttp.HandlerOpts{
		Registry: reg,
	}))

	server := &http.Server{
		Addr: addr,
	}

	go func() {
		<-ctx.Done()
		server.Close()
	}()

	return server.ListenAndServe()
}
