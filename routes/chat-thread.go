package routes

import (
	"net/http"
	"retro-chat-rooms/chatroom"

	"github.com/gin-gonic/gin"
)

func GetChatThread(c *gin.Context) {
	id := c.Param("id")
	userId := c.Param("userId")

	room := chatroom.FindRoomByID(id)

	if room == nil {
		c.Status(http.StatusNotFound)
		return
	}

	c.HTML(http.StatusOK, "chat-thread.html", gin.H{
		"ID":       room.ID,
		"Name":     room.Name,
		"Color":    room.Color,
		"UserID":   userId,
		"Messages": room.Messages,
	})
}
