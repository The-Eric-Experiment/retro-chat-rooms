package routes

import (
	"net/http"
	"retro-chat-rooms/chat"
	"retro-chat-rooms/config"

	"github.com/gin-gonic/gin"
)

func GetChatHeader(c *gin.Context) {
	roomId := c.Param("id")

	room, found := chat.GetSingleRoom(roomId)

	if !found {
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
