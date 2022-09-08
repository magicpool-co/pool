package common

import (
	"time"
)

func NormalizeDate(date time.Time, period time.Duration, backwards bool) time.Time {
	date = date.UTC()
	d, h, m, s := date.Day(), date.Hour(), date.Minute(), date.Second()

	switch period := period; {
	case period >= (time.Hour * 24):
		value := int(period) / int(time.Hour*24)
		d = processDateComponent(date.Day(), value, backwards)
		h, m, s = 0, 0, 0
	case period >= time.Hour:
		value := int(period) / int(time.Hour)
		h = processDateComponent(date.Hour(), value, backwards)
		m, s = 0, 0
	case period >= time.Minute:
		value := int(period) / int(time.Minute)
		m = processDateComponent(date.Minute(), value, backwards)
		s = 0
	case period >= time.Second:
		value := int(period) / int(time.Second)
		s = processDateComponent(date.Second(), value, backwards)
	}

	newDate := time.Date(date.Year(), date.Month(), d, h, m, s, 0, time.UTC)

	return newDate
}

func processDateComponent(component, value int, backwards bool) int {
	offset := component % value

	if backwards {
		return component - offset
	} else {
		return component + (value - offset)
	}
}
