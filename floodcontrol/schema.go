package floodcontrol

import "time"

type FloodControl struct {
	LastMessageFlood  time.Time
	MessageFloodCount int
	LastMaximumFlood  time.Time
}
