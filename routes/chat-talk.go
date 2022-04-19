package routes

import (
	"net/http"
	"retro-chat-rooms/chatroom"
	"retro-chat-rooms/session"

	"github.com/gin-gonic/gin"
)

func GetChatTalk(c *gin.Context) {
	id := c.Param("id")
	toUserId := c.Query("to")

	room := chatroom.FindRoomByID(id)

	if room == nil {
		c.Status(http.StatusNotFound)
		return
	}

	ident := session.GetSessionUserIdent(c)
	user := room.GetUserBySessionIdent(ident)

	if user == nil {
		c.Status(http.StatusNotFound)
		return
	}

	to := room.GetUser(toUserId)

	c.HTML(http.StatusOK, "chat-talk.html", gin.H{
		"ID":          room.ID,
		"UserId":      user.ID,
		"Nickname":    user.Nickname,
		"To":          to,
		"Color":       room.Color,
		"TextColor":   room.TextColor,
		"SpeechModes": chatroom.SPEECH_MODES,
	})
}
