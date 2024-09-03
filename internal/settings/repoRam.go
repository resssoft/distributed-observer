package settings

import (
	"fmt"
	"sync"

	models "observer/internal/domain/mediator"
	"observer/internal/domain/repository"
	"observer/pkg/requestFilter"
)

type repo struct {
	storage   map[string]models.SettingsItem
	mapSafety *sync.Mutex
}

func NewSettingsRepo() repository.Settings {
	return &repo{
		storage:   make(map[string]models.SettingsItem),
		mapSafety: &sync.Mutex{},
	}
}

func (r *repo) get(name string) (models.SettingsItem, error) {
	if item, ok := r.storage[name]; ok {
		return item, nil
	}
	return models.SettingsItem{}, fmt.Errorf("not found for name %v", name)
}

func (r *repo) GetList(filter requestFilter.Filter) ([]models.SettingsItem, error) {
	r.mapSafety.Lock()
	defer r.mapSafety.Unlock()
	name := ""
	for _, filterItem := range filter.Filters {
		for key, value := range filterItem.Data {
			if key == "name" {
				name = fmt.Sprintf("%v", &value)
			}
		}
	}
	founded, err := r.get(name)
	return []models.SettingsItem{founded}, err
}

func (r *repo) Update(item models.SettingsItem) (models.SettingsItem, error) {
	r.mapSafety.Lock()
	defer r.mapSafety.Unlock()
	_, err := r.get(item.Name)
	if err != nil {
		return item, err
	}
	r.storage[item.Name] = item
	return item, nil
}
