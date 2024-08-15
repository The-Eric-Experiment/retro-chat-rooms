package routes

import (
	"net"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type SessionUserState struct {
	session sessions.Session
	ctx     *gin.Context
	errors  []string
}

func NewSessionUserState(ctx *gin.Context, session sessions.Session) SessionUserState {
	errors := make([]string, 0)
	return SessionUserState{session, ctx, errors}
}

func (sus *SessionUserState) GetSupportsChatEventAwaiter() bool {
	supportsChatEventAwaiter := sus.session.Get("supportsChatEventAwaiter")

	if supportsChatEventAwaiter == nil {
		supportsChatEventAwaiter = true
	}

	return (supportsChatEventAwaiter.(bool))
}

func (sus *SessionUserState) GetUserID() string {
	return sus.session.Get("userId").(string)
}

func (sus *SessionUserState) GetCoolDownMessageSent() bool {
	coolDownMessageSent := sus.session.Get("coolDownMessageSent")

	if coolDownMessageSent == nil {
		coolDownMessageSent = false
	}

	return coolDownMessageSent.(bool)
}

func (sus *SessionUserState) SetCoolDownMessageSent(v bool) {
	sus.session.Set("coolDownMessageSent", v)
	sus.session.Save()
}

func (sus *SessionUserState) GetLastScream() time.Time {
	lastScream := sus.session.Get("lastScream")
	if lastScream == nil {
		return time.Unix(0, 0).UTC()
	}

	// Convert the retrieved value to an int64, assuming it's stored as such
	lastScreamUnix, ok := lastScream.(int64)
	if !ok {
		return time.Unix(0, 0).UTC() // Return Unix epoch time if there's a type mismatch
	}

	// Convert Unix time to time.Time
	return time.Unix(lastScreamUnix, 0).UTC()
}

func (sus *SessionUserState) SetLastScream(t time.Time) {
	// Convert time.Time to Unix time (int64)
	sus.session.Set("lastScream", t.Unix())
	sus.session.Save()
}

func (sus *SessionUserState) GetUserIP() string {
	headers := [4]string{
		"HTTP_CF_CONNECTING_IP", "HTTP_X_REAL_IP", "HTTP_X_FORWARDED_FOR", "REMOTE_ADDR",
	}

	for _, headerKey := range headers {
		ip := sus.ctx.Request.Header.Get(headerKey)
		if ip != "" {
			return ip
		}
	}

	ip, _, err := net.SplitHostPort(sus.ctx.Request.RemoteAddr)
	if err != nil {
		return ""
	}

	return ip
}

func (sus *SessionUserState) ValidateUser() {

}
