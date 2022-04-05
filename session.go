package main

import (
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

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

type UserSession map[string]interface{}

type Session struct {
	sessions map[string]UserSession
	mu       sync.Mutex
}

func (s *Session) Initialize() {
	s.sessions = make(map[string]UserSession)
}

func (s *Session) GetSessionValue(c *gin.Context, key string) (interface{}, bool) {
	defer s.mu.Unlock()
	s.mu.Lock()
	userIdent := GetSessionUserIdent(c)
	session := s.sessions[userIdent]

	if session[key] == nil {
		return nil, false
	}

	return session[key], true
}

func (s *Session) SetSessionValue(c *gin.Context, key string, value interface{}) {
	defer s.mu.Unlock()
	s.mu.Lock()
	userIdent := GetSessionUserIdent(c)
	if s.sessions[userIdent] == nil {
		s.sessions[userIdent] = UserSession{}
	}

	s.sessions[userIdent][key] = value
}

func (s *Session) DeregisterSession(c *gin.Context) {
	defer s.mu.Unlock()
	s.mu.Lock()
	userIdent := GetSessionUserIdent(c)
	delete(s.sessions, userIdent)
}
