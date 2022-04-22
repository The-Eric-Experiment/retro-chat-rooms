package routes

import (
	"net/http"
	"retro-chat-rooms/chatroom"
	"time"

	"github.com/gin-gonic/gin"
)

func GetChatUsers(c *gin.Context) {
	id := c.Param("id")

	room := chatroom.FindRoomByID(id)

	if room == nil {
		c.Status(http.StatusNotFound)
		return
	}

	user := room.GetUser(c)

	if user == nil {
		c.Status(http.StatusNotFound)
		return
	}

	c.HTML(http.StatusOK, "chat-users.html", gin.H{
		"ID":          room.ID,
		"UserID":      user.ID,
		"Users":       room.Users,
		"UsersOnline": room.GetOnlineUsers(),
		"RoomTime":    time.Now().UTC(),
	})
}
