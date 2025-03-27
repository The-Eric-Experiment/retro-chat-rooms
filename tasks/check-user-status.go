package tasks

import (
	"retro-chat-rooms/chat"
	"time"
)

func CheckUserStatus() {
	for {
		users := chat.GetAllUsers()
		for _, user := range users {
			if user.Client.Plat == chat.CLIENT_PLATFORM_WEB && chat.IsUserStale(user.ID) {
				chat.DeregisterUser(user.ID)
			}
		}

		time.Sleep(20000 * time.Millisecond)
	}
}
