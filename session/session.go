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

func GetSessionValue(c *gin.Context, key string) (interface{}, bool) {
	defer mu.Unlock()
	mu.Lock()
	userIdent := GetSessionUserIdent(c)
	session := sessions[userIdent]

	if session[key] == nil {
		return nil, false
	}

	return session[key], true
}

func SetSessionValue(c *gin.Context, key string, value interface{}) {
	defer mu.Unlock()
	mu.Lock()
	userIdent := GetSessionUserIdent(c)
	if sessions[userIdent] == nil {
		sessions[userIdent] = UserSession{}
	}

	sessions[userIdent][key] = value
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
