package routes

import (
	"net/http"
	"retro-chat-rooms/chatroom"
	"retro-chat-rooms/session"
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

	ident := session.GetSessionUserIdent(c)
	user := room.GetUserBySessionIdent(ident)

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
