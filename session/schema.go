package session

import "time"

type UserSession map[string]interface{}

type FloodControl struct {
	LastMessageFlood  time.Time
	MessageFloodCount int
	LastMaximumFlood  time.Time
}
