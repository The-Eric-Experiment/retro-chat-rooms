package floodcontrol

import (
	"sync"
	"time"
)

var ips = map[string]*FloodControl{}
var mu = sync.Mutex{}

func IsIPBanned(ip string) bool {

	mu.Lock()
	control := ips[ip]
	mu.Unlock()

	if control == nil {
		return false
	}

	return time.Now().UTC().Sub(control.LastMaximumFlood).Minutes() <= MESSAGE_FLOOD_IP_BAN_MIN
}

func IsCooldownPeriod(ip string) bool {
	mu.Lock()
	control := ips[ip]
	mu.Unlock()

	if control == nil {
		return false
	}

	return time.Now().UTC().Sub(control.LastMessageFlood).Minutes() <= MESSAGE_FLOOD_COOLDOWN_MIN
}

func SetFlooded(ip string) {
	defer mu.Unlock()
	mu.Lock()
	control := ips[ip]

	if control == nil {
		control = &FloodControl{}
		ips[ip] = control
	}

	now := time.Now().UTC()

	if control.MessageFloodCount >= MESSAGE_FLOOD_MAXIMUM_FLOODS {
		control.MessageFloodCount = 0
		control.LastMaximumFlood = now
	}

	control.LastMessageFlood = now
	control.MessageFloodCount++
}
