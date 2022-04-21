package routes

import (
	"net/http"
	"retro-chat-rooms/chatroom"

	"github.com/gin-gonic/gin"
)

func PostLogout(c *gin.Context) {
	id := c.PostForm("id")
	userId := c.PostForm("userId")

	room := chatroom.FindRoomByID(id)

	if room == nil {
		c.Status(http.StatusNotFound)
		return
	}

	user := room.GetUser(userId)

	if user == nil {
		c.Status(http.StatusNotFound)
		return
	}

	room.DeregisterUser(user)

	c.Redirect(301, "/chat-updater/"+room.ID)
}
