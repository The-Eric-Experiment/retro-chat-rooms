package tasks

import (
	"retro-chat-rooms/chatroom"
	"time"
)

func CheckUserStatus(rooms []*chatroom.Room) {
	for {
		for _, room := range rooms {
			// Check if there are expired users
			for _, user := range room.Users {
				if !user.IsDiscordUser && time.Now().UTC().Sub(user.LastPing).Seconds() > chatroom.USER_LOGOUT_TIMEOUT {
					room.DeregisterUser(user)
				}
			}
		}
		time.Sleep(20000 * time.Millisecond)
	}
}
