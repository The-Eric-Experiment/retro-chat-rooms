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
		c.Redirect(301, UrlJoin(room.ID))
		return
	}

	c.HTML(http.StatusOK, "room.html", gin.H{
		"Name":         room.Name,
		"Color":        room.Color,
		"ID":           room.ID,
		"HeaderHeight": config.Current.ChatRoomHeaderHeight,
	})
}
