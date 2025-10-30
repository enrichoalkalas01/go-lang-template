package server

import "context"

type Server interface {
	SetupMiddlewares()
	SetupRoutes()
	Start() error
	Shutdown(ctx context.Context) error
}
