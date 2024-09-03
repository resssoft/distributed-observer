package settings

import (
	"context"
	"encoding/json"
	"strconv"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"

	models "observer/internal/domain/mediator"
	"observer/internal/domain/repository"
	"observer/internal/domain/services"
	"observer/internal/logger"
	"observer/pkg/mediator"
	"observer/pkg/requestFilter"
)

const (
	eventsBuffer       = 10000
	cacheValueDuration = time.Minute * 5
)

type settingsData struct {
	repo       repository.Settings
	mapSafety  *sync.Mutex
	dispatcher *mediator.Dispatcher
	cache      *cache.Cache
	logger     *logger.Logger
}

var _ = (services.Settings)(&settingsData{})

func New(dispatcher *mediator.Dispatcher, logger *logger.Logger) services.Settings {
	logger.With("service", "settings")
	app := &settingsData{
		mapSafety:  &sync.Mutex{},
		dispatcher: dispatcher,
		cache:      cache.New(10*time.Minute, 20*time.Minute),
		logger:     logger,
		repo:       NewSettingsRepo(),
	}

	listener := Listener{
		Client: app,
		events: make(chan interface{}, eventsBuffer),
	}
	if err := dispatcher.Register(
		listener,
		models.SettingsEvents...); err != nil {
		logger.Error(context.Background(), err, "dispatcher.Register")
	}
	go app.eventsHandler(listener)
	go app.eventsHandler(listener)
	return app
}

func (r *settingsData) Validate(data []byte) (models.SettingsItem, error) {
	item := models.SettingsItem{}
	err := json.Unmarshal(data, &item)
	return item, err
}

func (r *settingsData) GetList(filter requestFilter.Filter) ([]models.SettingsItem, error) {
	return r.repo.GetList(filter)
}

func (r *settingsData) Update(item models.SettingsItem) (models.SettingsItem, error) {
	return r.repo.Update(item)
}

func (r *settingsData) GetValue(name, defaultVal string) string {
	if v, found := r.cache.Get(name); found {
		if value, converted := v.(string); converted {
			return value
		}
	}
	filter := requestFilter.GetSimpleFilter("=", "Name", name)
	values, err := r.GetList(filter)
	if err != nil {
		r.cache.Set(name, defaultVal, cacheValueDuration)
		return defaultVal
	}
	if len(values) > 0 {
		r.cache.Set(name, values[0].Value, cacheValueDuration)
		return values[0].Value
	}
	r.cache.Set(name, defaultVal, cacheValueDuration)
	return defaultVal
}

func (r *settingsData) GetValueInt(name string, defaultVal int) int {
	value := r.GetValue(name, strconv.Itoa(defaultVal))
	converted, _ := strconv.Atoi(value)
	return converted
}

func (r *settingsData) GetValueBool(name string, defaultVal bool) bool {
	defaultValStr := "false"
	if defaultVal {
		defaultValStr = "true"
	}
	value := r.GetValue(name, defaultValStr)
	return value == "1" || value == "true"
}

func (r *settingsData) GetValueSeconds(name string, defaultVal int) time.Duration {
	return time.Duration(r.GetValueInt(name, defaultVal)) * time.Second
}

func (r *settingsData) GetValueMinutes(name string, defaultVal int) time.Duration {
	return time.Duration(r.GetValueInt(name, defaultVal)) * time.Minute
}

func (r *settingsData) GetValueHours(name string, defaultVal int) time.Duration {
	return time.Duration(r.GetValueInt(name, defaultVal)) * time.Hour
}

func (r *settingsData) GetValueDays(name string, defaultVal int) time.Duration {
	return time.Duration(r.GetValueInt(name, defaultVal)) * time.Hour * 24
}

func (r *settingsData) SleepSecondsAt(name string, defaultVal int) {
	time.Sleep(time.Duration(r.GetValueInt(name, defaultVal)) * time.Second)
}

func (r *settingsData) AfterSeconds(name string, defaultVal int) <-chan time.Time {
	return time.After(r.GetValueSeconds(name, defaultVal))
}

func (r *settingsData) AfterMinutes(name string, defaultVal int) <-chan time.Time {
	return time.After(r.GetValueMinutes(name, defaultVal))
}
