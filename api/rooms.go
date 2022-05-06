package api

import (
	"net/http"
	"retro-chat-rooms/chat"

	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
)

func GetRooms(ctx *gin.Context) {
	rooms := chat.GetAllRooms()

	responseRooms := lo.Map(rooms, func(room chat.ChatRoom, _ int) ChatRoom {
		return ChatRoom{
			ID:     room.ID,
			Name:   room.Name,
			Color:  room.Color,
			Online: len(chat.GetRoomOnlineUsers(room.ID)),
		}
	})

	ctx.IndentedJSON(http.StatusOK, responseRooms)
}
