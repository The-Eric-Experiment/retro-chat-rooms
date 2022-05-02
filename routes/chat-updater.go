package routes

import (
	"net/http"
	"retro-chat-rooms/chat"
	"retro-chat-rooms/floodcontrol"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func unsub(roomId string, userId string) {
	chat.RoomListEvents[roomId].Unsubscribe(userId)
	chat.RoomMessageEvents[roomId].Unsubscribe(userId)
}

func waitForChatEvent(roomId string, userId string, cb func(userListEvent bool, userList bool)) {
	unsub(roomId, userId)
	select {
	case <-chat.RoomMessageEvents[roomId].Subscribe(userId):
		unsub(roomId, userId)
		cb(true, false)
		break
	case <-chat.RoomListEvents[roomId].Subscribe(userId):
		unsub(roomId, userId)
		cb(false, true)
		break
	case <-time.After(chat.UPDATER_WAIT_TIMEOUT_MS * time.Millisecond):
		unsub(roomId, userId)
		cb(false, false)
		break
	}
}

func getData(room chat.ChatRoom, userId string, hasMessages bool, userListUpdated bool, supportsAwaiter bool) gin.H {
	_, found := chat.GetUser(userId)

	if !found {
		return gin.H{
			"UserGone": true,
			"ID":       room.ID,
		}
	}

	chat.Ping(userId)

	return gin.H{
		"ID":                       room.ID,
		"HasMessages":              hasMessages,
		"UserListUpdated":          userListUpdated,
		"Color":                    room.Color,
		"SupportsChatEventAwaiter": supportsAwaiter,
	}
}

func GetChatUpdater(c *gin.Context, session sessions.Session) {
	roomId := c.Param("id")
	userId := session.Get("userId").(string)
	combinedId := chat.GetCombinedId(roomId, userId)
	room, found := chat.GetSingleRoom(roomId)

	if !found {
		c.Status(http.StatusNotFound)
		return
	}

	if floodcontrol.IsIPBanned(c) {
		chat.SendMessage(roomId, &chat.ChatMessage{
			Time:            time.Now().UTC(),
			Message:         "{nickname} was kicked for flooding the channel too many times.",
			IsSystemMessage: true,
			SpeechMode:      chat.MODE_SAY_TO,
		})

		chat.DeregisterUser(combinedId)

		c.HTML(http.StatusOK, "chat-updater.html", getData(room, combinedId, false, false, false))
		return
	}

	supportsChatEventAwaiter := session.Get("supportsChatEventAwaiter")

	if supportsChatEventAwaiter == nil {
		supportsChatEventAwaiter = true
	}

	if !supportsChatEventAwaiter.(bool) {
		hasMessages := chat.UserHasMessages(combinedId)
		userListUpdated := chat.HasUserListChanged(combinedId)

		data := getData(room, combinedId, hasMessages, userListUpdated, supportsChatEventAwaiter.(bool))

		c.HTML(http.StatusOK, "chat-updater.html", data)
		return
	}

	result := make(chan gin.H)

	go waitForChatEvent(roomId, combinedId, func(hasMessages bool, userListUpdated bool) {
		result <- getData(room, combinedId, hasMessages, userListUpdated, supportsChatEventAwaiter.(bool))
	})

	c.HTML(http.StatusOK, "chat-updater.html", <-result)
}
