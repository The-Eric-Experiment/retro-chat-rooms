package floodcontrol

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var ips = map[string]*FloodControl{}
var mu = sync.Mutex{}

func getIP(c *gin.Context) string {
	headers := [4]string{
		"HTTP_CF_CONNECTING_IP", "HTTP_X_REAL_IP", "HTTP_X_FORWARDED_FOR", "REMOTE_ADDR",
	}

	for _, headerKey := range headers {
		ip := c.Request.Header.Get(headerKey)
		if ip != "" {
			return ip
		}
	}

	return ""
}

func IsIPBanned(c *gin.Context) bool {
	ip := getIP(c)

	mu.Lock()
	control := ips[ip]
	mu.Unlock()

	if control == nil {
		return false
	}

	return time.Now().UTC().Sub(control.LastMaximumFlood).Minutes() <= MESSAGE_FLOOD_IP_BAN_MIN
}

func IsCooldownPeriod(c *gin.Context) bool {
	ip := getIP(c)

	mu.Lock()
	control := ips[ip]
	mu.Unlock()

	if control == nil {
		return false
	}

	return time.Now().UTC().Sub(control.LastMessageFlood).Minutes() <= MESSAGE_FLOOD_COOLDOWN_MIN
}

func SetFlooded(c *gin.Context) {
	ip := getIP(c)

	defer mu.Unlock()
	mu.Lock()
	control := ips[ip]

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
