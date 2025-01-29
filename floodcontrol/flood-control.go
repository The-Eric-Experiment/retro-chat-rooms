package floodcontrol

import (
	"sync"
	"time"
)

var (
	mu  sync.Mutex
	ips = map[string]*FloodControl{}
)

func getControl(ip string) *FloodControl {
	control, ok := ips[ip]
	if !ok {
		control = &FloodControl{}
		ips[ip] = control
	}
	return control
}

// IsIPBanned returns true if the IP is in a ban period.
func IsIPBanned(ip string) bool {
	mu.Lock()
	defer mu.Unlock()

	c := getControl(ip)
	// If current time is before the ban’s end, it’s banned
	return time.Now().Before(c.BanEndTime)
}

// IsCooldownPeriod returns true if the IP is still in cooldown
// (even if not fully banned).
func IsCooldownPeriod(ip string) bool {
	mu.Lock()
	defer mu.Unlock()

	c := getControl(ip)
	// If current time is before NextAllowedMessage, it’s on cooldown
	return time.Now().Before(c.NextAllowedMessage)
}

// RecordMessage should be called whenever the user sends a message.
// This function updates the IP’s sliding window, checks for flooding,
// and, if needed, sets cooldown or triggers a ban.
func RecordMessage(ip string) {
	mu.Lock()
	defer mu.Unlock()

	c := getControl(ip)

	now := time.Now()

	// If we’re past NextAllowedMessage, it means cooldown has expired
	if now.After(c.NextAllowedMessage) {
		c.NextAllowedMessage = time.Time{}
	}

	// 1) If currently banned, do nothing else.
	// 2) If currently in cooldown, do nothing else.
	// The chat code can still reject the message, but we keep track anyway:
	// because even if they're in cooldown, they might spam attempts that count
	// toward further bans. For your usage, you might want to handle differently.

	// Clean up old timestamps from the sliding window
	cutoff := now.Add(-time.Duration(MESSAGE_FLOOD_RANGE_SEC) * time.Second)
	i := 0
	for i < len(c.RecentMessages) && c.RecentMessages[i].Before(cutoff) {
		i++
	}
	c.RecentMessages = c.RecentMessages[i:] // drop old messages

	// Add this message to the list
	c.RecentMessages = append(c.RecentMessages, now)

	// Check if it’s a flood
	if len(c.RecentMessages) >= (MESSAGE_FLOOD_MESSAGES_COUNT + 1) {
		// We triggered a flood → set cooldown

		// Only increment FloodCount if we’re not currently in cooldown
		if c.NextAllowedMessage.IsZero() || now.After(c.NextAllowedMessage) {
			c.FloodCount++
		}

		// If user’s floods exceed maximum, ban them
		if c.FloodCount >= MESSAGE_FLOOD_MAXIMUM_FLOODS {
			c.BanEndTime = now.Add(time.Duration(MESSAGE_FLOOD_IP_BAN_MIN) * time.Minute)
			c.FloodCount = 0
		}

		// Always set the new cooldown window
		c.NextAllowedMessage = now.Add(time.Duration(MESSAGE_FLOOD_COOLDOWN_SEC) * time.Second)
	}
}
