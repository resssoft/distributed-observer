package models

import (
	"observer/pkg/mediator"
)

const SettingsItemSave mediator.EventName = "settings.item.save"

var SettingsEvents = []mediator.EventName{
	SettingsItemSave,
}

type SettingsEvent struct {
	Item SettingsItem
}

type SettingsItem struct {
	Id          int    `db:"id"`
	Name        string `db:"name"`
	Value       string `db:"value"`
	Group       string `db:"group"`
	Type        string `db:"type"`
	Data        string `db:"data"`
	UserId      int    `db:"user_id"`
	Title       string `db:"title"`
	Description string `db:"description"`
}

const (
	DateTimeMicroFormat = "2006-01-02--15:04:05.000000"
	CacheKeyPrefix      = "all:vmm:items:"
	DateTimeUtcFormat   = "2006-01-02 15:04:05 -0700"
)
