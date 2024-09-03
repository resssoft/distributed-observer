package repository

import (
	models "observer/internal/domain/mediator"
	"observer/pkg/requestFilter"
)

type Settings interface {
	GetList(requestFilter.Filter) ([]models.SettingsItem, error)
	Update(models.SettingsItem) (models.SettingsItem, error)
}
