package routes

import (
	"net/http"
	"retro-chat-rooms/chat"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func GetChatThread(c *gin.Context, session sessions.Session) {
	roomId := c.Param("id")

	userId := session.Get("userId")
	combinedId := chat.GetCombinedId(roomId, userId.(string))

	messages, found := chat.GetUserMessageList(combinedId)

	if !found {
		c.Status(http.StatusNotFound)
		return
	}

	c.HTML(http.StatusOK, "chat-thread.html", gin.H{
		"UserID":   combinedId,
		"Messages": messages,
	})
}
