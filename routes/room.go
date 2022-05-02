package routes

import (
	"net/http"
	"retro-chat-rooms/chat"
	"retro-chat-rooms/config"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func GetRoom(c *gin.Context, session sessions.Session) {
	roomId := c.Param("id")
	room, found := chat.GetSingleRoom(roomId)
	if !found {
		c.Status(http.StatusNotFound)
		return
	}

	userId := session.Get("userId")
	combinedId := chat.GetCombinedId(roomId, userId.(string))
	_, found = chat.GetUser(combinedId)

	if !found {
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
