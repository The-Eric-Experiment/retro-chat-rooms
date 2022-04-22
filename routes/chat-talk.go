package routes

import (
	"net/http"
	"regexp"
	"retro-chat-rooms/chatroom"
	"retro-chat-rooms/floodcontrol"
	"retro-chat-rooms/profanity"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
)

func checkForMessageAbuse(room *chatroom.Room, user *chatroom.RoomUser) bool {
	if user.IsDiscordUser || user.IsAdmin {
		return false
	}

	messages := room.GetUserMessages(user)

	if len(messages) < floodcontrol.MESSAGE_FLOOD_MESSAGES_COUNT {
		return false
	}

	firstMessage, err := lo.Last(messages[0:floodcontrol.MESSAGE_FLOOD_MESSAGES_COUNT])
	if err != nil {
		return false
	}

	messages = messages[0 : floodcontrol.MESSAGE_FLOOD_MESSAGES_COUNT-1]

	for _, msg := range messages {
		seconds := msg.Time.Sub(firstMessage.Time).Seconds()
		if seconds > floodcontrol.MESSAGE_FLOOD_RANGE_SEC {
			return false
		}
	}

	return true
}

func sendHtml(c *gin.Context, room *chatroom.Room, user *chatroom.RoomUser, toUserId string, updateUpdater bool) {
	to := room.GetUserByID(toUserId)

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

	user := room.GetUser(c)

	if user == nil {
		c.Status(http.StatusNotFound)
		return
	}

	sendHtml(c, room, user, toUserId, false)
}

func PostChatTalk(c *gin.Context, session sessions.Session) {
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

	user := room.GetUser(c)

	if user == nil {
		c.Status(http.StatusNotFound)
		return
	}

	if len(strings.TrimSpace(message)) == 0 {
		sendHtml(c, room, user, toUserId, false)
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
		sendHtml(c, room, user, toUserId, updateUpdater)
		return
	}

	coolDownMessageSent := session.Get("coolDownMessageSent")

	if floodcontrol.IsCooldownPeriod(c) && (coolDownMessageSent == nil || !coolDownMessageSent.(bool)) {
		room.SendMessage(&chatroom.RoomMessage{
			Time:            now,
			To:              user,
			From:            user,
			IsSystemMessage: true,
			Message:         "Hey {nickname}, chill out, you'll be able to send messages again in " + strconv.FormatInt(int64(floodcontrol.MESSAGE_FLOOD_COOLDOWN_MIN), 10) + " minutes.",
			Privately:       true,
			SpeechMode:      chatroom.MODE_SAY_TO,
		})
		session.Set("coolDownMessageSent", true)
		session.Save()
		sendHtml(c, room, user, toUserId, updateUpdater)
		return
	}

	session.Set("coolDownMessageSent", false)
	session.Save()

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

	lastScream := session.Get("lastScream")

	if lastScream != nil && now.Sub(*lastScream.(*time.Time)).Minutes() <= chatroom.USER_SCREAM_TIMEOUT_MIN && mode == chatroom.MODE_SCREAM_AT {
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

	userTo := room.GetUserByID(to)

	if mode == chatroom.MODE_SCREAM_AT {
		session.Set("lastScream", &now)
		session.Save()
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

	user := room.GetUserByID(m.Author.ID)
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
