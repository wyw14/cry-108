package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wyw14/fabchem/internal/api"
)

func main() {
	listen := flag.String("listen", "127.0.0.1:21208", "HTTP listen address")
	data := flag.String("data", "./data", "local persistent data directory")
	flag.Parse()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	runtime, err := api.NewRuntime(*data)
	if err != nil {
		logger.Error("runtime initialization failed", "error", err)
		os.Exit(1)
	}
	server := &http.Server{
		Addr: *listen, Handler: api.NewServer(runtime, logger).Handler(),
		ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 20 * time.Second,
		WriteTimeout: 20 * time.Second, IdleTimeout: 60 * time.Second,
	}
	stopping, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-stopping.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = runtime.Snapshot(ctx)
		_ = server.Shutdown(ctx)
	}()
	logger.Info("fabchem listening", "address", *listen, "data", *data)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("HTTP server failed", "error", err)
		os.Exit(1)
	}
}
