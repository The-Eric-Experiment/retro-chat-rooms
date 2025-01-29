package sockets

import (
	"time"
)

type SocketsUserState struct {
	conn ISocket
}

func NewSocketsUserState(conn ISocket) SocketsUserState {
	return SocketsUserState{conn}
}

func (sus *SocketsUserState) GetUserID() string {
	return sus.conn.GetUser().ID
}

func (sus *SocketsUserState) GetCoolDownMessageSent() bool {
	coolDownMessageSent := sus.conn.GetState("coolDownMessageSent")

	if coolDownMessageSent == nil {
		coolDownMessageSent = false
	}

	return coolDownMessageSent.(bool)
}

func (sus *SocketsUserState) SetCoolDownMessageSent(v bool) {
	sus.conn.SetState("coolDownMessageSent", v)
}

func (sus *SocketsUserState) GetLastScream() time.Time {
	lastScream := sus.conn.GetState("lastScream")
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

func (sus *SocketsUserState) SetLastScream(t time.Time) {
	// Convert time.Time to Unix time (int64)
	sus.conn.SetState("lastScream", t.Unix())
}

func (sus *SocketsUserState) GetUserIP() string {
	return sus.conn.GetClientIP()
}
