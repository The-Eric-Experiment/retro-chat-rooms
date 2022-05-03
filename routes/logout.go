package routes

import (
	"net/http"
	"retro-chat-rooms/chat"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func PostLogout(c *gin.Context, session sessions.Session) {
	roomId := c.PostForm("id")

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

	suportsChatEventAwaiter := session.Get("supportsChatEventAwaiter")

	if suportsChatEventAwaiter == nil {
		suportsChatEventAwaiter = true
	}

	c.HTML(http.StatusOK, "logout.html", gin.H{
		"SupportsEventAwaiter": suportsChatEventAwaiter,
		"ID":                   roomId,
	})
}
