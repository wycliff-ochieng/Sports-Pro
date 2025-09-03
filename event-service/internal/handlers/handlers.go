package handlers

import (
	"log"

	"github.com/wycliff-ochieng/internal/service"
)

type EventHandler struct {
	logger log.Logger
	es     *service.EventService
}

func NewEventHandler(l log.Logger, es *service.EventService) *EventHandler {
	return &EventHandler{
		logger: l,
		es:     es,
	}
}
