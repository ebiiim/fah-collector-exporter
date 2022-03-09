package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [-addr] [-insecure] COLLECTOR_VIEWER_URL\n", os.Args[0])
		flag.PrintDefaults()
	}

	var addr string
	flag.StringVar(&addr, "addr", ":8080", "The address to listen on for HTTP requests")
	var insecure bool
	flag.BoolVar(&insecure, "insecure", false, "Skip verifying collector's TLS cert")

	flag.Parse()
	if len(flag.Args()) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	url := flag.Arg(0)
	exp := NewExporter(url, insecure)

	prometheus.MustRegister(exp)

	http.Handle("/healthz", http.HandlerFunc(
		func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(addr, nil))
}
