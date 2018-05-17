package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"fmt"

	"github.com/dave/frizz/config"
	"github.com/dave/frizz/server/assets"
	"github.com/dave/frizz/server/handler"
)

func init() {
	assets.Init()
}

func main() {

	var svr *http.Server

	shutdown := make(chan struct{})
	hdl := handler.New(shutdown)

	if config.DEV {
		svr = &http.Server{Addr: ":" + fmt.Sprint(config.DevServerPort), Handler: hdl}
	} else {
		// In GKE we default to port 8080 and use the value of the "PORT" environment variable if set.
		port := "8080"
		if fromEnv := os.Getenv("PORT"); fromEnv != "" {
			port = fromEnv
		}
		svr = &http.Server{Addr: ":" + port, Handler: hdl}
	}

	go func() {
		log.Print("Listening on " + svr.Addr)
		if err := svr.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	// Set up graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Wait for shutdown signal
	<-stop

	// Signal to all the compile handlers that the server wants to shut down
	close(shutdown)

	ctx, cancel := context.WithTimeout(context.Background(), config.ServerShutdownTimeout)
	defer cancel()

	// Wait for all compile jobs to be cancelled
	hdl.Waitgroup.Wait()

	if err := svr.Shutdown(ctx); err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		log.Println("Main server stopped")
	}
}
