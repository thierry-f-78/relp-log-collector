package main

import "flag"
import "fmt"
import "log/slog"
import "os"

import "github.com/thierry-f-78/relp-log-collector/pkg/backend"
import "github.com/thierry-f-78/relp-log-collector/pkg/config"
import "github.com/thierry-f-78/relp-log-collector/pkg/dispatch"
import "github.com/thierry-f-78/relp-log-collector/pkg/relp"
import "github.com/thierry-f-78/relp-log-collector/pkg/utilities"

func main() {
	var err error
	var configFile string
	var log *slog.Logger

	// Configure program flags and parse command line
	flag.StringVar(&configFile, "f", "", "Configuration file")
	flag.Parse()

	err = config.Load(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}

	// Init log system
	log = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{}))

	// Load backend plugins
	err = backend.Init(log)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}

	// Init RELP server
	err = relp.InitRELPServer()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}

	// inform systemd
	err = utilities.NotifySystemd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}

	// Start server processes
	go relp.StartRELPServer(log)
	dispatch.Dispatch(log)
}
