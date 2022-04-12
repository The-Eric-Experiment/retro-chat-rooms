package session

import (
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/samber/lo"
)

var sessions = map[string]UserSession{}
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

func GetSessionUserIdent(ctx *gin.Context) string {
	uagent := ctx.Request.Header.Get("User-Agent")
	ip := getIP(ctx)
	return uuid.NewSHA1(uuid.NameSpaceURL, []byte(ip+uagent)).String()
}

func GetSessionValue(sessionIdent string, key string) (interface{}, bool) {
	defer mu.Unlock()
	mu.Lock()
	session := sessions[sessionIdent]

	if session[key] == nil {
		return nil, false
	}

	return session[key], true
}

func SetSessionValue(sessionIdent string, key string, value interface{}) {
	defer mu.Unlock()
	mu.Lock()
	if sessions[sessionIdent] == nil {
		sessions[sessionIdent] = UserSession{}
	}

	sessions[sessionIdent][key] = value
}

func DeregisterSession(userIdent string) {
	defer mu.Unlock()
	mu.Lock()
	delete(sessions, userIdent)
}

func GetSessionIds() []string {
	defer mu.Unlock()
	mu.Lock()
	return lo.Keys(sessions)
}

func RegisterUserIP(c *gin.Context) string {
	ip := getIP(c)
	registerIp(ip)
	sessionIdent := GetSessionUserIdent(c)
	SetSessionValue(sessionIdent, "ip", ip)
	return ip
}
