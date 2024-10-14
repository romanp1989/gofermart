package server

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/romanp1989/gophermart/internal/config"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type handlerFunc func(mux *chi.Mux)

type App struct {
	server   *http.Server
	log      *zap.Logger
	handlers []handlerFunc
}

func NewApp(log *zap.Logger, router *chi.Mux) *App {
	log.Info("Running server on ", zap.String("port", config.Options.FlagServerAddress))

	return &App{
		server: &http.Server{
			Addr:    config.Options.FlagServerAddress,
			Handler: router,
		},
	}
}

func (app *App) RunServer() error {
	errChannel := make(chan error, 1)

	go func() {
		err := app.server.ListenAndServe()
		if err != nil {
			errChannel <- err
			return
		}

		close(errChannel)
	}()

	return <-errChannel
}

func (app *App) Stop() {
	ctx, shutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdown()

	_ = app.server.Shutdown(ctx)
}
