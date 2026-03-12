package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"sgroups.io/sgroups-k8s-api/internal/backend"
	"sgroups.io/sgroups-k8s-api/internal/mock"
	"sgroups.io/sgroups-k8s-api/internal/seed"
)

func main() {
	addr := flag.String("addr", ":8081", "gRPC listen address")
	seedPath := flag.String("seed", "", "Path to JSON seed file (optional)")
	flag.Parse()

	mb := mock.New()
	if *seedPath != "" {
		if err := seed.Load(*seedPath, mb, mb); err != nil {
			log.Fatalf("seed load failed: %v", err)
		}
	}

	b := backend.Backend{
		Namespaces:      mb,
		AddressGroups:   mb,
		Networks:        mb,
		Hosts:           mb,
		HostBindings:    mb,
		NetworkBindings: mb,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Printf("sgroups-k8s-api listening on %s", *addr)
	if err := mock.Run(ctx, *addr, b); err != nil {
		log.Fatalf("server stopped: %v", err) //nolint:gocritic // exitAfterDefer: stop() cleanup is not critical at process exit
	}
}
