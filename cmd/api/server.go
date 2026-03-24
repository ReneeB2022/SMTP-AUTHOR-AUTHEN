package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (a *applicationDependencies) serve() error {
	apiServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", a.config.port),
		Handler:      a.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(a.logger.Handler(), slog.LevelError),
	}
	shutdownError := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)                      // receive the shutdown signal
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM) // signal occurred
		s := <-quit                                          // blocks until a signal is received
		// message about shutdown in process
		a.logger.Info("shutting down server", "signal", s.String())
		// create a context
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		shutdownError <- apiServer.Shutdown(ctx)
	}()

	a.logger.Info("starting server", "address", apiServer.Addr,
		"environment", a.config.environment)

	err := apiServer.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return apiServer.ListenAndServe()

}
