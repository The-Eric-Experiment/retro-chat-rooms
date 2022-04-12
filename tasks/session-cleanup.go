package tasks

import (
	"retro-chat-rooms/chatroom"
	"retro-chat-rooms/session"
	"time"

	"github.com/samber/lo"
)

func SessionCleanup(rooms []*chatroom.Room) {
	for {
		sessionIds := session.GetSessionIds()

		sessionIds = lo.Filter(sessionIds, func(sessionId string, _ int) bool {
			for _, room := range rooms {
				for _, user := range room.Users {
					if user.SessionIdent == sessionId {
						return false
					}
				}
			}

			return true
		})

		for _, id := range sessionIds {
			session.DeregisterSession(id)
		}

		time.Sleep(120000 * time.Millisecond)
	}
}
