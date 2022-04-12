package routes

import (
	"fmt"
	"net/http"
	"retro-chat-rooms/chatroom"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
)

func GetChatUpdater(c *gin.Context) {
	id := c.Param("id")
	userId := c.Param("userId")

	room := chatroom.FindRoomByID(id)

	if room == nil {
		c.Status(http.StatusNotFound)
		return
	}

	user := room.GetUser(userId)

	if user == nil {
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

	fmt.Println(userListUpdated)

	user.LastPing = time.Now().UTC()

	getData := func() gin.H {
		return gin.H{
			"ID":              room.ID,
			"UserID":          user.ID,
			"HasMessages":     hasMessages,
			"UserListUpdated": userListUpdated,
			"Color":           room.Color,
		}
	}

	if userListUpdated || hasMessages {
		c.HTML(http.StatusOK, "chat-updater.html", getData())
		return
	}

	result := make(chan gin.H)

	room.ChatEventAwaiter.Await(c, func(context *gin.Context) {
		result <- getData()
	})

	c.HTML(http.StatusOK, "chat-updater.html", <-result)
}
