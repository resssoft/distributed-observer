package manager

import (
	"context"

	"observer/internal/domain/services"
	"observer/internal/logger"
	"observer/internal/pinger"
	"observer/internal/settings"
	"observer/pkg/mediator"
)

type Data struct {
	dispatcher *mediator.Dispatcher
	logger     *logger.Logger
	Services   Services
}

type Services struct {
	settings services.Settings
	pinger   *pinger.Data
}

var onExit chan bool

func New() *Data {
	dispatcher := mediator.NewDispatcher()
	loggerService := logger.New(nil, nil)
	settingsService := settings.New(dispatcher, loggerService)
	return &Data{
		dispatcher: dispatcher,
		logger:     loggerService,
		Services: Services{
			settings: settingsService,
			pinger:   pinger.New(dispatcher, loggerService, settingsService),
		},
	}
}

func (d *Data) Start(ctx context.Context) {
	println("start manager")
	d.Services.pinger.Start(ctx)
	<-onExit
}
