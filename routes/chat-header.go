package routes

import (
	"net/http"
	"retro-chat-rooms/chatroom"
	"retro-chat-rooms/config"

	"github.com/gin-gonic/gin"
)

func GetChatHeader(c *gin.Context) {
	id := c.Param("id")

	room := chatroom.FindRoomByID(id)

	if room == nil {
		c.Status(http.StatusNotFound)
		return
	}

	c.HTML(http.StatusOK, "chat-header.html", gin.H{
		"Name":      room.Name,
		"Color":     room.Color,
		"TextColor": room.TextColor,
		"Logo":      config.Current.ChatRoomHeaderLogo,
	})
}
