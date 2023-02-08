package main

import (
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/urfave/cli/v2"
)

func maybeStartProfiler(c *cli.Context) error {
	bindIPAndPort := c.String(flagProfile)
	if bindIPAndPort == "" {
		return nil
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	server := &http.Server{
		Addr:              bindIPAndPort,
		ReadHeaderTimeout: time.Second,
		Handler:           mux,
	}

	go func() {
		err := server.ListenAndServe()
		log.Fatalf("Profiling server failed: %s", err)
	}()
	time.Sleep(100 * time.Millisecond) // let the profiler start

	// While some data is browsable directly via http://127.0.0.1:${YOUR_PORT}/debug/pprof,
	// other parts like cpu profiling need to be parsed using "pprof". Example:
	//   go tool pprof http://127.0.0.1:${YOUR_PORT}/debug/pprof/profile
	//   (pprof) top
	//      [shows top CPU using functions]
	//   (pprof) list FUNCTION_NAME_FROM_ABOVE_OUTPUT
	//      [shows cpu of line numbers in function]
	log.Infof("Serving pprof data (browsable): http://%s/debug/pprof", bindIPAndPort)
	return nil
}
