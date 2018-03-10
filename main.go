package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/olivere/elastic"
)

var (
	port      = 8080
	seedCount = 1000
	sgnl      = make(chan os.Signal, 1) // register a channel to listen for system signals
)

func init() {
	// register a channel to listen for system signals
	signal.Notify(sgnl, syscall.SIGINT, syscall.SIGTERM)
}

func main() {

	/////////////////////////
	//     connections     //
	/////////////////////////

	// memStore := NewMemStore("host")

	client, err := elastic.NewClient(
		elastic.SetURL("http://127.0.0.1:9200"),
		elastic.SetSniff(false), // creates trouble if elasticsearch is running in docker and this app is running outside of Docker
	)
	if err != nil {
		panic(err)
	}
	defer client.Stop()

	esFetcher := NewElasticStore(client)

	// setup indexes before starting anything
	if err := esFetcher.SetupIndex(); err != nil {
		panic(fmt.Sprintf("Unable to setup index! Error: %v", err))
	}

	// setup indexes before starting anything
	if err := esFetcher.SeedData(seedCount); err != nil {
		panic(fmt.Sprintf("Unable to seed data! Error: %v", err))
	}

	fmt.Print("Successfully setup index and seed data")

	/////////////////////////
	//     app creation    //
	/////////////////////////

	app := &app{fetcher: esFetcher}

	/////////////////////////
	//     server entry    //
	/////////////////////////

	http.Handle("/", app.hint())
	http.Handle("/search", app.search())

	s := &http.Server{Addr: ":" + strconv.Itoa(port)}

	// start shutdown sequence in separate goroutine
	go gracefulShutdown(s)

	log.Printf("Listening on port %v", port)
	if err := s.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("Listen and serve error: %v", err)
	}
}

// Separate goroutine to listen for interrupt signal. This will trigger the http server to shutdown
func gracefulShutdown(server *http.Server) {
	// block on system signal
	<-sgnl

	log.Printf("Received interrupt signal; shutting down server")

	// Note that this doesn't help with "fancy" connections like Websockets; those are up to us to close up
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	server.Shutdown(ctx)
}
