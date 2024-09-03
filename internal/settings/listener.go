package settings

import (
	"context"

	models "observer/internal/domain/mediator"
	"observer/pkg/mediator"
)

type Listener struct {
	Client *settingsData
	events chan interface{}
}

func (u Listener) Push(_ mediator.EventName, eventData interface{}) {
	u.events <- eventData
}

func (n *settingsData) eventsHandler(u Listener) {
	for event := range u.events {
		u.Listen("", event)
	}
}

func (u Listener) Listen(_ mediator.EventName, event interface{}) {
	switch event := event.(type) {
	case models.SettingsEvent:
		u.Client.Update(event.Item)
	default:
		u.Client.logger.Info(context.Background(), "registered an invalid notifier event", "event", event)
	}
}
