package routes

import (
	"net/http"
	"retro-chat-rooms/chat"

	"github.com/gin-gonic/gin"
)

func GetIndex(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"Rooms": chat.GetAllRooms(),
	})
}
