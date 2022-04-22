package routes

import (
	"net/http"
	"retro-chat-rooms/chatroom"

	"github.com/gin-gonic/gin"
)

func PostLogout(c *gin.Context) {
	id := c.PostForm("id")

	room := chatroom.FindRoomByID(id)

	if room == nil {
		c.Status(http.StatusNotFound)
		return
	}

	user := room.GetUser(c)

	if user == nil {
		c.Status(http.StatusNotFound)
		return
	}

	room.DeregisterUser(user)

	// TODO: This shouldn't happen here
	c.Redirect(301, UrlChatUpdater(room.ID))
}
