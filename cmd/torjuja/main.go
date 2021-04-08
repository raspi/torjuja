package main

import (
	"flag"
	"fmt"
	"github.com/raspi/torjuja/pkg/db/fsdb"
	"github.com/raspi/torjuja/pkg/db/iface"
	"github.com/raspi/torjuja/pkg/service"
	"os"
	"strings"
)

var (
	// These are set with Makefile -X=main.VERSION, etc
	VERSION   = `v0.0.0`
	BUILD     = `dev`
	BUILDDATE = `0000-00-00T00:00:00+00:00`
)

const (
	AUTHOR   = `Pekka Järvinen`
	HOMEPAGE = `https://github.com/raspi/torjuja`
	YEAR     = 2021
)

func main() {
	showVersionArg := flag.Bool(`version`, false, `Show version`)
	configArg := flag.String(`config`, `config.json`, `Configuration file`)

	flag.Usage = func() {
		_, _ = fmt.Fprintf(os.Stdout, `torjuja - DNS blocker`+"\n")
		_, _ = fmt.Fprintf(os.Stdout, `Version %v (%v)`+"\n", VERSION, BUILDDATE)
		_, _ = fmt.Fprintf(os.Stdout, `(c) %v %v- [ %v ]`+"\n", AUTHOR, YEAR, HOMEPAGE)
		_, _ = fmt.Fprintf(os.Stdout, "\n")

		_, _ = fmt.Fprintf(os.Stdout, "Parameters:\n")

		paramMaxLen := 0

		flag.VisitAll(func(f *flag.Flag) {
			l := len(f.Name)
			if l > paramMaxLen {
				paramMaxLen = l
			}
		})

		flag.VisitAll(func(f *flag.Flag) {
			padding := strings.Repeat(` `, paramMaxLen-len(f.Name))
			_, _ = fmt.Fprintf(os.Stdout, "  -%s%s   %s   default: %q\n", f.Name, padding, f.Usage, f.DefValue)
		})

		_, _ = fmt.Fprintf(os.Stdout, "\n")
	}

	flag.Parse()

	if *showVersionArg {
		// Show version information
		_, _ = fmt.Fprintf(os.Stdout, `Version %s build %s built on %s`+"\n", VERSION, BUILD, BUILDDATE)
		os.Exit(0)
	}

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
