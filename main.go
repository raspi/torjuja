package main

import (
	"flag"
	"fmt"
	"github.com/raspi/torjuja/pkg/db/fsdb"
	"github.com/raspi/torjuja/pkg/db/iface"
	"github.com/raspi/torjuja/pkg/service"
	"os"
)

func main() {
	configArg := flag.String(`config`, `config.json`, `Configuration file`)
	flag.Parse()

	cfg, err := service.LoadConfig(*configArg)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, `error: %v`, err)
		os.Exit(1)
	}

	var db iface.Database

	db, err = fsdb.New(cfg.Database.FileSystem.Path)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, `error: %v`, err)
		os.Exit(1)
	}

	errs := make(chan error)
	defer close(errs)

	s, err := service.New(cfg.ListenAddresses, cfg.Blocked, cfg.TTL, cfg.Forwarders, db, cfg.ApiListen, errs)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, `error: %v`, err)
		os.Exit(1)
	}

	if err := s.Listen(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, `error: %v`, err)
		os.Exit(1)
	}

	fmt.Printf(`HTTP server http://%s`+"\n", cfg.ApiListen)
	for _, s := range cfg.ListenAddresses {
		fmt.Printf(`DNS server %s`+"\n", s)
	}

	fmt.Printf(`Sending blocked IPv4 to %q`+"\n", cfg.Blocked.IPv4)
	fmt.Printf(`Sending blocked IPv6 to %q`+"\n", cfg.Blocked.IPv6)
	fmt.Printf(`Sending blocked PTR to %q`+"\n", cfg.Blocked.PTR)
	fmt.Printf(`TTL: %d seconds`+"\n", cfg.TTL)

	for e := range errs {
		switch e.(type) {
		case service.StartupFailureError:
			_, _ = fmt.Fprintf(os.Stderr, `could not start server: %v`, e)
			os.Exit(1)
		}

		_, _ = fmt.Fprintf(os.Stderr, `error: %v`, e)
	}
}
