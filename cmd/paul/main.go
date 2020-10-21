package main

import (
	"context"
	"fmt"
	"github.com/Spazzy757/paul/pkg/helpers"
	"github.com/Spazzy757/paul/pkg/router"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const startUpLog = `
__________  _____   ____ ___.____     
\______   \/  _  \ |    |   \    |    
 |     ___/  /_\  \|    |   /    |    
 |    |  /    |    \    |  /|    |___ 
 |____|  \____|__  /______/ |_______ \
                 \/                 \/
`

func main() {
	// Termination Handeling
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)
	// Get the routes
	router := router.GetRouter()
	// Set server configuration
	port := helpers.GetEnv("SERVER_PORT", "8000")
	host := helpers.GetEnv("SERVER_HOST", "127.0.0.1")
	// String formatting to join the host and port
	addr := fmt.Sprintf("%v:%v", host, port)
	// Setup Server
	srv := &http.Server{
		Handler:      router,
		Addr:         addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	// Run Server in Goroutine to handle Graceful Shutdowns
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	fmt.Println(startUpLog)
	log.WithFields(log.Fields{
		"host": host,
		"port": port,
	}).Info("Starting Server")
	<-termChan
	// Any Code to Gracefully Shutdown should be done here
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}
	log.Println("Shutting Down Gracefully")
}
