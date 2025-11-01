package handler

import "service-golang/pkg/logger"

type UserHandler struct {
	Logger *logger.Logger
}

func NewUserHandler(logger *logger.Logger) *UserHandler {
	return &UserHandler{
		Logger: logger,
	}
}

func (h *UserHandler) Get() error {
	return nil
}
