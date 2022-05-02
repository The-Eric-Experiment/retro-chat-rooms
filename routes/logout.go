package routes

import (
	"net/http"
	"retro-chat-rooms/chat"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func PostLogout(c *gin.Context, session sessions.Session) {
	roomId := c.Param("id")
	room, found := chat.GetSingleRoom(roomId)
	if !found {
		c.Status(http.StatusNotFound)
		return
	}

	userId := session.Get("userId")
	combinedId := chat.GetCombinedId(roomId, userId.(string))
	user, found := chat.GetUser(combinedId)

	if !found {
		c.Status(http.StatusNotFound)
		return
	}

	if user.IsAdmin && user.IsDiscordUser() {
		session.Set("userId", nil)
		session.Save()
	} else {
		chat.DeregisterUser(combinedId)
	}

	// TODO: This shouldn't happen here
	c.Redirect(301, UrlChatUpdater(room.ID))
}
