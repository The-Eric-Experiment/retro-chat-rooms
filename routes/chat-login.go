package routes

import (
	"net/http"
	"retro-chat-rooms/chatroom"

	"github.com/gin-gonic/gin"
)

func GetChatLogin(c *gin.Context) {
	id := c.Param("id")

	room := chatroom.FindRoomByID(id)

	if room == nil {
		c.Status(http.StatusNotFound)
		return
	}

	c.HTML(http.StatusOK, "chat-login.html", gin.H{
		"Color":  room.Color,
		"Name":   room.Name,
		"Colors": chatroom.NICKNAME_COLORS,
		"ID":     room.ID,
	})
}
