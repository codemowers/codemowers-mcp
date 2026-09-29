package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
)

var version = "dev"

func main() {
	if err := run(); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	content, err := fs.Sub(assets, "content")
	if err != nil {
		return err
	}
	sections, err := loadAdvisories(content)
	if err != nil {
		return err
	}
	publicURL := os.Getenv("PUBLIC_URL")
	if publicURL == "" {
		publicURL = "https://mcp.codemowers.io"
	}
	handler, err := newHandler(sections, publicURL)
	if err != nil {
		return err
	}
	address := os.Getenv("HTTP_ADDR")
	if address == "" {
		address = "[::]:3000"
	}
	healthAddress := os.Getenv("HEALTH_ADDR")
	if healthAddress == "" {
		healthAddress = "[::]:3001"
	}
	var ready atomic.Bool
	health := http.NewServeMux()
	health.HandleFunc("GET /readyz", func(w http.ResponseWriter, _ *http.Request) {
		if !ready.Load() {
			http.Error(w, "not ready", http.StatusServiceUnavailable)
			return
		}
		_, _ = fmt.Fprintln(w, "ok")
	})
	health.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) { _, _ = fmt.Fprintln(w, "ok") })
	servers := []*http.Server{
		{Addr: address, Handler: handler, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second},
		{Addr: healthAddress, Handler: health, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 5 * time.Second, WriteTimeout: 5 * time.Second, IdleTimeout: 30 * time.Second},
	}
	var listeners []net.Listener
	defer func() {
		for _, listener := range listeners {
			_ = listener.Close()
		}
	}()
	for _, server := range servers {
		listener, err := net.Listen("tcp", server.Addr)
		if err != nil {
			return err
		}
		listeners = append(listeners, listener)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	failures := make(chan error, len(servers))
	for i, server := range servers {
		go func() {
			if err := server.Serve(listeners[i]); err != nil && !errors.Is(err, http.ErrServerClosed) {
				failures <- err
			}
		}()
	}
	ready.Store(true)
	slog.Info("serving advisory", "address", address, "publicURL", publicURL, "documents", len(sections))
	select {
	case <-ctx.Done():
	case err = <-failures:
	}
	ready.Store(false)
	shutdown, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	for _, server := range servers {
		if closeErr := server.Shutdown(shutdown); closeErr != nil {
			_ = server.Close()
			err = errors.Join(err, closeErr)
		}
	}
	return err
}
