package routes

import (
	"net/http"
	"retro-chat-rooms/chatroom"
	"retro-chat-rooms/floodcontrol"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
)

func GetChatUpdater(c *gin.Context, session sessions.Session) {
	id := c.Param("id")

	room := chatroom.FindRoomByID(id)

	if room == nil {
		c.Status(http.StatusNotFound)
		return
	}

	user := room.GetUser(c)

	if user == nil {
		c.HTML(http.StatusOK, "chat-updater.html", gin.H{
			"UserGone": true,
			"ID":       room.ID,
		})
		return
	}

	if floodcontrol.IsIPBanned(c) {
		room.SendMessage(&chatroom.RoomMessage{
			Time:            time.Now().UTC(),
			Message:         "{nickname} was kicked for flooding the channel too many times.",
			IsSystemMessage: true,
			SpeechMode:      chatroom.MODE_SAY_TO,
		})

		room.DeregisterUser(user)
		c.HTML(http.StatusOK, "chat-updater.html", gin.H{
			"UserGone": true,
			"ID":       room.ID,
		})
		return
	}

	_, hasMessages := lo.Find(
		room.Messages,
		func(m *chatroom.RoomMessage) bool { return m.Time.Sub(user.LastPing).Seconds() > 0 },
	)

	userListUpdated := room.LastUserListUpdate.Sub(user.LastUserListUpdate).Seconds() > 0

	if userListUpdated {
		user.LastUserListUpdate = time.Now().UTC()
	}

	user.LastPing = time.Now().UTC()

	supportsChatEventAwaiter := session.Get("supportsChatEventAwaiter")

	if supportsChatEventAwaiter == nil {
		supportsChatEventAwaiter = true
	}

	getData := func() gin.H {
		return gin.H{
			"ID":                       room.ID,
			"HasMessages":              hasMessages,
			"UserListUpdated":          userListUpdated,
			"Color":                    room.Color,
			"SupportsChatEventAwaiter": supportsChatEventAwaiter.(bool),
		}
	}

	if userListUpdated || hasMessages || !supportsChatEventAwaiter.(bool) {
		c.HTML(http.StatusOK, "chat-updater.html", getData())
		return
	}

	result := make(chan gin.H)

	room.ChatEventAwaiter.Await(c, func(context *gin.Context) {
		result <- getData()
	})

	c.HTML(http.StatusOK, "chat-updater.html", <-result)
}
