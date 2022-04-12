package routes

import (
	"net/http"
	"retro-chat-rooms/chatroom"
	"time"

	"github.com/gin-gonic/gin"
)

func GetChatUsers(c *gin.Context) {
	id := c.Param("id")
	userId := c.Param("userId")

	room := chatroom.FindRoomByID(id)

	if room == nil {
		c.Status(http.StatusNotFound)
		return
	}

	user := room.GetUser(userId)

	if user == nil {
		c.Status(http.StatusNotFound)
		return
	}

	c.HTML(http.StatusOK, "chat-users.html", gin.H{
		"ID":          room.ID,
		"UserID":      user.ID,
		"Users":       room.Users,
		"UsersOnline": len(room.Users),
		"RoomTime":    time.Now().UTC(),
		"ChatOwner":   chatroom.OwnerRoomUser,
	})
}
