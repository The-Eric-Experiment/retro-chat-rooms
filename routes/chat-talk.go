package routes

import (
	"net/http"
	"retro-chat-rooms/chat"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func sendHtml(ctx *gin.Context, room chat.ChatRoom, user chat.ChatUser, toUserId string, updateUpdater bool, private bool) {
	var to chat.ChatUser
	if toUserId != "" {
		toUser, fnd := chat.GetUser(toUserId)
		if fnd {
			to = toUser
		}
	}

	ctx.HTML(http.StatusOK, "chat-talk.html", gin.H{
		"ID":            room.ID,
		"UserId":        user.ID,
		"Nickname":      user.Nickname,
		"To":            to,
		"Color":         room.Color,
		"TextColor":     room.TextColor,
		"SpeechModes":   chat.SPEECH_MODES,
		"UpdateUpdater": updateUpdater,
		"Private":       private,
	})
}

func getSupportsChatEventAwaiter(session sessions.Session) bool {
	supportsChatEventAwaiter := session.Get("supportsChatEventAwaiter")

	if supportsChatEventAwaiter == nil {
		supportsChatEventAwaiter = true
	}

	return (supportsChatEventAwaiter.(bool))
}

func GetChatTalk(c *gin.Context, session sessions.Session) {
	roomId := c.Param("id")
	toUserId := c.Query("to")

	room, found := chat.GetSingleRoom(roomId)

	if !found {
		c.Status(http.StatusNotFound)
		return
	}

	userId := session.Get("userId")

	combinedId := chat.GetCombinedId(roomId, userId.(string))
	user, found := chat.GetUser(combinedId)

	if !found {
		c.Status(http.StatusNotFound)
		return
	}

	sendHtml(c, room, user, toUserId, false, false)
}

func PostChatTalk(c *gin.Context, session sessions.Session) {
	roomId := c.PostForm("id")
	message := c.PostForm("message")
	mode := c.PostForm("speechMode")
	private := c.PostForm("private")
	toUserId := c.PostForm("to")

	sessionUserState := NewSessionUserState(c, session)
	// TODO: Figure out why this needs ! and write a comment
	updateUpdater := !getSupportsChatEventAwaiter(session)

	room, found := chat.GetSingleRoom(roomId)

	if !found {
		c.Status(http.StatusNotFound)
		return
	}

	userId := sessionUserState.GetUserID()

	combinedId := chat.GetCombinedId(roomId, userId)
	user, found := chat.GetUser(combinedId)

	if !found {
		c.Status(http.StatusNotFound)
		return
	}

	involvedUsers := []chat.ChatUser{user}
	toUser, foundToUser := chat.GetUser(toUserId)

	if foundToUser {
		involvedUsers = append(involvedUsers, toUser)
	}

	finalMessage, canSend := chat.ValidateMessage(&sessionUserState, chat.ChatMessage{
		RoomID:          room.ID,
		Time:            time.Now().UTC(),
		Message:         message,
		From:            user.ID,
		To:              toUserId,
		Privately:       private == "on",
		SpeechMode:      mode,
		IsSystemMessage: false,
		Source:          chat.ClientInfoToMsgSource(user.Client),
		InvolvedUsers:   involvedUsers,
	})

	if !canSend {
		sendHtml(c, room, user, toUserId, false, false)
		return
	}

	chat.SendMessage(&finalMessage)

	sendHtml(c, room, user, toUserId, updateUpdater, private == "on")
}
