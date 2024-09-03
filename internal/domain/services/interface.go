package services

import (
	"time"
)

type Settings interface {
	GetValue(string, string) string
	GetValueInt(string, int) int
	GetValueBool(string, bool) bool
	GetValueSeconds(name string, defaultVal int) time.Duration
	GetValueMinutes(name string, defaultVal int) time.Duration
	GetValueHours(name string, defaultVal int) time.Duration
	GetValueDays(name string, defaultVal int) time.Duration
	SleepSecondsAt(name string, defaultVal int)
	AfterSeconds(name string, defaultVal int) <-chan time.Time
	AfterMinutes(name string, defaultVal int) <-chan time.Time
}
