package routes

import (
	"net/http"
	"retro-chat-rooms/chatroom"
	"retro-chat-rooms/config"
	"retro-chat-rooms/session"

	"github.com/gin-gonic/gin"
)

func GetRoom(c *gin.Context) {
	id := c.Param("id")

	room := chatroom.FindRoomByID(id)

	if room == nil {
		c.Status(http.StatusNotFound)
		return
	}

	sessionIdent := session.GetSessionUserIdent(c)
	user := room.GetUserBySessionIdent(sessionIdent)

	if user == nil {
		c.Redirect(http.StatusTemporaryRedirect, "/join/"+room.ID)
		return
	}

	c.HTML(http.StatusFound, "main.html", gin.H{
		"UserID":       user.ID,
		"Name":         room.Name,
		"Color":        room.Color,
		"ID":           room.ID,
		"Nickname":     user.Nickname,
		"HeaderHeight": config.Current.ChatRoomHeaderHeight,
	})
}
