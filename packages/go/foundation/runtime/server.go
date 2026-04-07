package runtime

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"
)

type HTTPConfig struct {
	Address           string
	ReadHeaderTimeout time.Duration
	ShutdownTimeout   time.Duration
}

func RunHTTP(ctx context.Context, logger *slog.Logger, config HTTPConfig, handler http.Handler) error {
	server := &http.Server{
		Addr:              config.Address,
		Handler:           handler,
		ReadHeaderTimeout: config.ReadHeaderTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("http server starting", "address", config.Address)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}

		errCh <- nil
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), config.ShutdownTimeout)
		defer cancel()

		logger.Info("http server shutting down")
		return server.Shutdown(shutdownCtx)
	}
}
