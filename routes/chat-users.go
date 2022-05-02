package routes

import (
	"net/http"
	"retro-chat-rooms/chat"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func GetChatUsers(c *gin.Context, session sessions.Session) {
	roomId := c.Param("id")

	room, found := chat.GetSingleRoom(roomId)

	if !found {
		c.Status(http.StatusNotFound)
		return
	}

	userId := session.Get("userId")
	combinedId := chat.GetCombinedId(roomId, userId.(string))
	user, found := chat.GetUser(combinedId)

	if !found {
		c.Status(http.StatusNotFound)
		return
	}

	onlineUsers := chat.GetRoomOnlineUsers(roomId)

	c.HTML(http.StatusOK, "chat-users.html", gin.H{
		"ID":          room.ID,
		"UserID":      user.ID,
		"Users":       onlineUsers,
		"UsersOnline": len(onlineUsers),
		"RoomTime":    time.Now().UTC(),
	})
}
