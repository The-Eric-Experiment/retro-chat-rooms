package routes

import (
	"net/http"
	"retro-chat-rooms/chatroom"
	"retro-chat-rooms/session"

	"github.com/gin-gonic/gin"
)

func GetChatThread(c *gin.Context) {
	id := c.Param("id")

	ident := session.GetSessionUserIdent(c)

	room := chatroom.FindRoomByID(id)

	if room == nil {
		c.Status(http.StatusNotFound)
		return
	}

	user := room.GetUserBySessionIdent(ident)

	if user == nil {
		c.Status(http.StatusNotFound)
		return
	}

	c.HTML(http.StatusOK, "chat-thread.html", gin.H{
		"ID":       room.ID,
		"Name":     room.Name,
		"Color":    room.Color,
		"UserID":   user.ID,
		"Messages": room.Messages,
	})
}
