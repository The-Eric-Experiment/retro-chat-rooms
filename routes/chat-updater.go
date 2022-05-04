package routes

import (
	"net/http"
	"retro-chat-rooms/chat"
	"retro-chat-rooms/floodcontrol"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func waitForChatEvent(roomId string, combinedId string, cb func(messageUpdates bool, userListUpdates bool)) {
	select {
	case val := <-chat.RoomEvents[roomId].Subscribe(combinedId):
		data := val.(chat.ChatEvent)
		cb(data.Message != nil, data.IsUserListUpdate)
		break
	case <-time.After(chat.UPDATER_WAIT_TIMEOUT_MS * time.Millisecond):
		cb(false, false)
		break
	}
}

func getData(room chat.ChatRoom, combinedId string, hasMessages bool, userListUpdated bool, supportsAwaiter bool) (gin.H, bool) {
	_, found := chat.GetUser(combinedId)

	if !found {
		return gin.H{
			"UserGone": true,
			"ID":       room.ID,
		}, true
	}

	chat.Ping(combinedId)

	return gin.H{
		"ID":                       room.ID,
		"HasMessages":              hasMessages,
		"UserListUpdated":          userListUpdated,
		"Color":                    room.Color,
		"SupportsChatEventAwaiter": supportsAwaiter,
	}, !supportsAwaiter || hasMessages || userListUpdated
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

		data, _ := getData(room, combinedId, false, false, false)

		c.HTML(http.StatusOK, "chat-updater.html", data)
		return
	}

	supportsChatEventAwaiter := session.Get("supportsChatEventAwaiter")

	if supportsChatEventAwaiter == nil {
		supportsChatEventAwaiter = true
	}

	hasMessages := chat.UserHasMessages(combinedId)
	userListUpdated := chat.HasUserListChanged(combinedId)

	data, updated := getData(room, combinedId, hasMessages, userListUpdated, supportsChatEventAwaiter.(bool))

	// If it has changes, return straight away.
	if updated {
		c.HTML(http.StatusOK, "chat-updater.html", data)
		return
	}

	result := make(chan gin.H)

	go waitForChatEvent(roomId, combinedId, func(hasMessages bool, userListUpdated bool) {
		data, _ := getData(room, combinedId, hasMessages, userListUpdated, supportsChatEventAwaiter.(bool))
		result <- data
	})

	c.HTML(http.StatusOK, "chat-updater.html", <-result)
}
