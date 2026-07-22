package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"

	"github.com/itchio/mitch"
)

func main() {
	port := flag.Int("port", 0, "port to listen on")
	flag.IntVar(port, "p", 0, "port to listen on (shorthand)")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	s, err := mitch.NewServer(ctx, mitch.WithPort(*port))
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Now listening on %s", s.Address())
	log.Printf("(Ctrl+C to exit)")
	<-ctx.Done()
}
