package session

import (
	"time"
)

var ips = map[string]*FloodControl{}

func IsIPBanned(sessionIdent string) bool {
	ip, found := GetSessionValue(sessionIdent, "ip")
	if !found {
		return false
	}
	mu.Lock()
	control := ips[ip.(string)]
	mu.Unlock()

	if control == nil {
		return false
	}

	return time.Now().UTC().Sub(control.LastMaximumFlood).Minutes() <= MESSAGE_FLOOD_IP_BAN_MIN
}

func IsCooldownPeriod(sessionIdent string) bool {
	ip, found := GetSessionValue(sessionIdent, "ip")
	if !found {
		return false
	}
	mu.Lock()
	control := ips[ip.(string)]
	mu.Unlock()

	if control == nil {
		return false
	}

	return time.Now().UTC().Sub(control.LastMessageFlood).Minutes() <= MESSAGE_FLOOD_COOLDOWN_MIN
}

func SetFlooded(sessionIdent string) {
	ip, found := GetSessionValue(sessionIdent, "ip")
	if !found {
		return
	}
	defer mu.Unlock()
	mu.Lock()
	control := ips[ip.(string)]

	if control == nil {
		return
	}

	now := time.Now().UTC()

	if control.MessageFloodCount >= MESSAGE_FLOOD_MAXIMUM_FLOODS {
		control.MessageFloodCount = 0
		control.LastMaximumFlood = now
	}

	control.LastMessageFlood = now
	control.MessageFloodCount++
}

func registerIp(ip string) {
	defer mu.Unlock()
	mu.Lock()
	control := ips[ip]

	if control == nil {
		ips[ip] = &FloodControl{}
	}
}
