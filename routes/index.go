package routes

import (
	"net/http"
	"retro-chat-rooms/chatroom"

	"github.com/gin-gonic/gin"
)

func GetIndex(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"Rooms": chatroom.CHAT_ROOMS,
	})
}
