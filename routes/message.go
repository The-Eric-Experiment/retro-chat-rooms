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

func PostMessage(c *gin.Context) {
	id := c.PostForm("id")
	message := c.PostForm("message")
	userId := c.PostForm("userId")
	to := c.PostForm("to")
	mode := c.PostForm("speechMode")
	private := c.PostForm("private")
	now := time.Now().UTC()

	if len(strings.TrimSpace(message)) == 0 {
		c.Redirect(302, "/chat-updater/"+id+"/"+userId)
		return
	}

	room := chatroom.FindRoomByID(id)

	if room == nil {
		c.Status(http.StatusNotFound)
		return
	}

	user := room.GetUser(userId)

	if user == nil {
		c.Status(http.StatusNotFound)
		return
	}

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

		c.Redirect(302, "/chat-updater/"+room.ID+"/"+user.ID)
		return
	}

	message = profanity.ReplaceSensoredProfanity(message)

	// Check if user has screamed recently

	lastScream, hasValue := session.GetSessionValue(c, "lastScream")

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

		c.Redirect(302, "/chat-updater/"+room.ID+"/"+user.ID)
		return
	}

	userTo := room.GetUser(to)

	if mode == chatroom.MODE_SCREAM_AT {
		session.SetSessionValue(c, "lastScream", &now)
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

	c.Redirect(302, "/chat-updater/"+room.ID+"/"+user.ID)
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
