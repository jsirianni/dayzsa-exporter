// Package main is the entry point of the application.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	signalCtx, signalCancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer signalCancel()

	for {
		ticker := time.NewTicker(10 * time.Second)
		select {
		case <-ticker.C:
			fmt.Println("I'm doing something !")
		case <-signalCtx.Done():
			fmt.Println("Shutting down...")
			cancel()
			os.Exit(1)
		}
	}
}
