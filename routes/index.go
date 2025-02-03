package routes

import (
	"net/http"
	"retro-chat-rooms/chat"

	"github.com/gin-gonic/gin"
)

func GetIndex(c *gin.Context) {
	rooms := chat.GetAllRooms()

	send := gin.H{
		"Rooms": rooms,
	}

	c.HTML(http.StatusOK, "index.html", send)
}
