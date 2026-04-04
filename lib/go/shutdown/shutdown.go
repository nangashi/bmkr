// Package shutdown provides a graceful HTTP server shutdown helper that
// waits for OS signals and coordinates teardown.
package shutdown

import "net/http"

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
// wip: use signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
// and select on serveErr and ctx.Done().
// On non-ErrServerClosed serveErr: log error, call cleanup, os.Exit(1).
// On ctx.Done(): proceed to graceful shutdown.
// Create a 10-second shutdown context, call server.Shutdown.
// On timeout: log, call server.Close, call cleanup, os.Exit(1).
// On clean shutdown: log "server stopped", call cleanup.
//
// Parameters:
//   - server:   the running *http.Server.
//   - serveErr: buffered channel (cap >= 1) that receives the error from
//     server.ListenAndServe.
//   - cleanup:  called exactly once before return or os.Exit; must not be nil.
func GracefulShutdown(server *http.Server, serveErr <-chan error, cleanup func()) {
	panic("not implemented")
}
