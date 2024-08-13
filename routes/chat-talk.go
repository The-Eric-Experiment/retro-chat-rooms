package routes

import (
	"net/http"
	"retro-chat-rooms/chat"
	"retro-chat-rooms/floodcontrol"
	"retro-chat-rooms/profanity"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func checkForMessageAbuse(room chat.ChatRoom, user chat.ChatUser) bool {
	if user.IsDiscordUser() || user.IsAdmin {
		return false
	}

	messages, found := chat.GetMessagesByUser(user.ID)

	if !found {
		return false
	}

	if len(messages) < floodcontrol.MESSAGE_FLOOD_MESSAGES_COUNT {
		return false
	}

	lastMessage := messages[len(messages)-1]

	initial := len(messages) - floodcontrol.MESSAGE_FLOOD_MESSAGES_COUNT

	messages = messages[initial : len(messages)-1]

	for _, msg := range messages {
		seconds := msg.Time.Sub(lastMessage.Time).Seconds()
		if seconds > floodcontrol.MESSAGE_FLOOD_RANGE_SEC {
			return false
		}
	}

	return true
}

func sendHtml(c *gin.Context, room chat.ChatRoom, user chat.ChatUser, toUserId string, updateUpdater bool, private bool) {
	var to chat.ChatUser
	if toUserId != "" {
		toUser, fnd := chat.GetUser(toUserId)
		if fnd {
			to = toUser
		}
	}

	c.HTML(http.StatusOK, "chat-talk.html", gin.H{
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
	now := time.Now().UTC()

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

	if len(strings.TrimSpace(message)) == 0 {
		sendHtml(c, room, user, toUserId, false, private == "on")
		return
	}

	suportsChatEventAwaiter := session.Get("supportsChatEventAwaiter")

	if suportsChatEventAwaiter == nil {
		suportsChatEventAwaiter = true
	}

	updateUpdater := !(suportsChatEventAwaiter.(bool))

	if checkForMessageAbuse(room, user) {
		floodcontrol.SetFlooded(c)
	}

	if floodcontrol.IsIPBanned(c) {
		sendHtml(c, room, user, toUserId, updateUpdater, private == "on")
		return
	}

	coolDownMessageSent := session.Get("coolDownMessageSent")

	if floodcontrol.IsCooldownPeriod(c) && (coolDownMessageSent == nil || !coolDownMessageSent.(bool)) {
		chat.SendMessage(roomId, &chat.ChatMessage{
			Time:                 now,
			To:                   combinedId,
			IsSystemMessage:      true,
			Message:              "Hey {nickname}, chill out, you'll be able to send messages again in " + strconv.FormatInt(int64(floodcontrol.MESSAGE_FLOOD_COOLDOWN_MIN), 10) + " minutes.",
			Privately:            true,
			SystemMessageSubject: &user,
			SpeechMode:           chat.MODE_SAY_TO,
			InvolvedUsers:        []chat.ChatUser{user},
		})

		session.Set("coolDownMessageSent", true)
		session.Save()
		sendHtml(c, room, user, toUserId, updateUpdater, private == "on")
		return
	}

	session.Set("coolDownMessageSent", false)
	session.Save()

	// Check if there's slurs

	if profanity.HasBlockedWords(message) {
		chat.SendMessage(roomId, &chat.ChatMessage{
			Time:                 now,
			To:                   combinedId,
			IsSystemMessage:      true,
			Message:              "Come on {nickname}! Let's be nice! This is a place for having fun!",
			Privately:            true,
			SystemMessageSubject: &user,
			SpeechMode:           chat.MODE_SAY_TO,
			InvolvedUsers:        []chat.ChatUser{user},
		})

		sendHtml(c, room, user, toUserId, updateUpdater, private == "on")
		return
	}

	message = profanity.ReplaceSensoredProfanity(message)

	// Check if user has screamed recently

	lastScream := session.Get("lastScream")

	if lastScream != nil && now.Sub(*lastScream.(*time.Time)).Minutes() <= chat.USER_SCREAM_TIMEOUT_MIN && mode == chat.MODE_SCREAM_AT {
		chat.SendMessage(roomId, &chat.ChatMessage{
			Time:                 now,
			To:                   combinedId,
			IsSystemMessage:      true,
			SystemMessageSubject: &user,
			Message:              "Hi {nickname}, you're only allowed to scream once very " + strconv.FormatInt(int64(chat.USER_SCREAM_TIMEOUT_MIN), 10) + " minutes.",
			Privately:            true,
			SpeechMode:           chat.MODE_SAY_TO,
			InvolvedUsers:        []chat.ChatUser{user},
		})

		sendHtml(c, room, user, toUserId, updateUpdater, private == "on")
		return
	}

	if mode == chat.MODE_SCREAM_AT {
		session.Set("lastScream", now)
		session.Save()
	}

	involvedUsers := []chat.ChatUser{user}
	toUser, foundToUser := chat.GetUser(toUserId)

	if foundToUser {
		involvedUsers = append(involvedUsers, toUser)
	}

	chat.SendMessage(roomId, &chat.ChatMessage{
		Time:            now,
		Message:         message,
		From:            combinedId,
		To:              toUserId,
		IsSystemMessage: false,
		Privately:       private == "on",
		SpeechMode:      mode,
		InvolvedUsers:   involvedUsers,
	})

	sendHtml(c, room, user, toUserId, updateUpdater, private == "on")
}
