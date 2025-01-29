package floodcontrol

import "time"

type FloodControl struct {
	// Timestamps for the last few messages (sliding window)
	RecentMessages     []time.Time
	NextAllowedMessage time.Time // If in cooldown, user can’t talk until this
	BanEndTime         time.Time // If banned, user can’t talk until this
	FloodCount         int       // How many times this IP has triggered a flood
}
