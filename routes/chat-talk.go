package routes

import (
	"net/http"
	"regexp"
	"retro-chat-rooms/chatroom"
	"retro-chat-rooms/profanity"
	"retro-chat-rooms/session"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gin-gonic/gin"
)

func sendHtml(c *gin.Context, room *chatroom.Room, user *chatroom.RoomUser, toUserId string, updateUpdater bool) {
	to := room.GetUser(toUserId)

	c.HTML(http.StatusOK, "chat-talk.html", gin.H{
		"ID":            room.ID,
		"UserId":        user.ID,
		"Nickname":      user.Nickname,
		"To":            to,
		"Color":         room.Color,
		"TextColor":     room.TextColor,
		"SpeechModes":   chatroom.SPEECH_MODES,
		"UpdateUpdater": updateUpdater,
	})
}

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

	sendHtml(c, room, user, toUserId, false)
}

func PostChatTalk(c *gin.Context) {
	id := c.PostForm("id")
	message := c.PostForm("message")
	to := c.PostForm("to")
	mode := c.PostForm("speechMode")
	private := c.PostForm("private")
	toUserId := c.PostForm("to")
	now := time.Now().UTC()

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

	if len(strings.TrimSpace(message)) == 0 {
		sendHtml(c, room, user, toUserId, false)
		return
	}

	suportsChatEventAwaiter, found := session.GetSessionValue(ident, "supportsChatEventAwaiter")

	if !found {
		suportsChatEventAwaiter = true
	}

	updateUpdater := !(suportsChatEventAwaiter.(bool))

	if session.IsIPBanned(user.SessionIdent) {
		sendHtml(c, room, user, toUserId, updateUpdater)
		return
	}

	coolDownMessageSent, found := session.GetSessionValue(user.SessionIdent, "coolDownMessageSent")

	if session.IsCooldownPeriod(user.SessionIdent) && (!found || !coolDownMessageSent.(bool)) {
		room.SendMessage(&chatroom.RoomMessage{
			Time:            now,
			To:              user,
			From:            user,
			IsSystemMessage: true,
			Message:         "Hey {nickname}, chill out, you'll be able to send messages again in " + strconv.FormatInt(int64(session.MESSAGE_FLOOD_COOLDOWN_MIN), 10) + " minutes.",
			Privately:       true,
			SpeechMode:      chatroom.MODE_SAY_TO,
		})
		session.SetSessionValue(user.SessionIdent, "coolDownMessageSent", true)
		sendHtml(c, room, user, toUserId, updateUpdater)
		return
	}

	session.SetSessionValue(user.SessionIdent, "coolDownMessageSent", false)

	// Check if there's slurs

	if profanity.HasBlockedWords(message) {
		room.SendMessage(&chatroom.RoomMessage{
			Time:            now,
			To:              user,
			From:            user,
			IsSystemMessage: true,
			Message:         "Come on {nickname}! Let's be nice! This is a place for having fun!",
			Privately:       true,
			SpeechMode:      chatroom.MODE_SAY_TO,
		})

		sendHtml(c, room, user, toUserId, updateUpdater)
		return
	}

	message = profanity.ReplaceSensoredProfanity(message)

	// Check if user has screamed recently

	lastScream, hasValue := session.GetSessionValue(user.SessionIdent, "lastScream")

	if hasValue && now.Sub(*lastScream.(*time.Time)).Minutes() <= chatroom.USER_SCREAM_TIMEOUT_MIN && mode == chatroom.MODE_SCREAM_AT {
		room.SendMessage(&chatroom.RoomMessage{
			Time:            now,
			To:              user,
			From:            user,
			IsSystemMessage: true,
			Message:         "Hi {nickname}, you're only allowed to scream once very " + strconv.FormatInt(int64(chatroom.USER_SCREAM_TIMEOUT_MIN), 10) + " minutes.",
			Privately:       true,
			SpeechMode:      chatroom.MODE_SAY_TO,
		})

		sendHtml(c, room, user, toUserId, updateUpdater)
		return
	}

	userTo := room.GetUser(to)

	if mode == chatroom.MODE_SCREAM_AT {
		session.SetSessionValue(user.SessionIdent, "lastScream", &now)
	}

	room.SendMessage(&chatroom.RoomMessage{
		Time:            now,
		Message:         message,
		From:            user,
		To:              userTo,
		IsSystemMessage: false,
		Privately:       private == "on",
		SpeechMode:      mode,
	})

	room.Update()

	sendHtml(c, room, user, toUserId, updateUpdater)
}

func ReceiveDiscordMessage(m *discordgo.MessageCreate) {
	content := m.Content

	room := chatroom.FindRoomByDiscordChannel(m.ChannelID)

	if room == nil {
		return
	}

	user := room.GetUser(m.Author.ID)
	if user == nil {
		user = room.RegisterDiscordUser(m.Author)
	}
	now := time.Now().UTC()

	messageMentionExpr := regexp.MustCompile(`^\s*@([^:]+):\s*(.+)`)

	match := messageMentionExpr.FindStringSubmatch(m.Content)

	var to *chatroom.RoomUser

	if len(match) > 1 {
		content = match[2]
		to = room.GetUserByNickname(match[1])
	}

	room.SendMessage(&chatroom.RoomMessage{
		Time:            now,
		Message:         content,
		From:            user,
		SpeechMode:      chatroom.MODE_SAY_TO,
		To:              to,
		IsSystemMessage: false,
		FromDiscord:     true,
	})

	room.Update()
}
