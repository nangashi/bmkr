// Package shutdown provides a graceful HTTP server shutdown helper that
// waits for OS signals and coordinates teardown.
package shutdown

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// GracefulShutdown blocks until the server exits gracefully or a fatal
// condition is encountered.
//
// Call sequence expected by the caller in main.go:
//
//	server := &http.Server{...}
//	serveErr := make(chan error, 1)
//	go func() {
//	    slog.Info("<service> service starting", "port", port)
//	    serveErr <- server.ListenAndServe()
//	}()
//	shutdown.GracefulShutdown(server, serveErr, pool.Close)
//
// Parameters:
//   - server:   the running *http.Server.
//   - serveErr: buffered channel (cap >= 1) that receives the error from
//     server.ListenAndServe.
//   - cleanup:  called exactly once before return or os.Exit; must not be nil.
func GracefulShutdown(server *http.Server, serveErr <-chan error, cleanup func()) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-serveErr:
		if !errors.Is(err, http.ErrServerClosed) {
			slog.Error("failed to start server", "error", err)
			cleanup()
			os.Exit(1)
		}
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	slog.Info("shutting down server")
	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("graceful shutdown timed out, forcing close", "error", err)
		_ = server.Close()
		cleanup()
		os.Exit(1)
	}
	cleanup()
	slog.Info("server stopped")
}
