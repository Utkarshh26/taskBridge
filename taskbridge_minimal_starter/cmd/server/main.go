package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"taskbridge/internal/server"
	"taskbridge/internal/store"
)

func main() {
	addr := flag.String("addr", ":8080", "server listen address")
	flag.Parse()
	st := store.NewMemoryStore()
	srv := server.New(st)

	stop := make(chan struct{})
	go srv.RunTimeoutWatcher(stop, 2*time.Second)

	httpServer := &http.Server{
		Addr:    *addr,
		Handler: srv.Handler(),
	}
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		close(stop)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(ctx)
	}()
	fmt.Printf("TaskBridge server listening on %s\n", *addr)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
