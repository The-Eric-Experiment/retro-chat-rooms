package routes

import (
	"net/http"
	"retro-chat-rooms/chat"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func GetIndex(c *gin.Context) {
	rooms := chat.GetAllRooms()

	// TODO: Add setting and config = Join is Index:
	joinData := getJoinData(rooms[0], "/", make([]string, 0))

	send := gin.H{
		"Rooms": rooms,
	}

	for k, v := range *joinData {
		send[k] = v
	}

	c.HTML(http.StatusOK, "index.html", send)
}

func PostIndex(c *gin.Context, session sessions.Session) {
	rooms := chat.GetAllRooms()

	failure := validateAndJoin(c, session, rooms[0], "/")

	if failure == nil {
		return
	}

	c.HTML(http.StatusOK, "index.html", failure)
}
