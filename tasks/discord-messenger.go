package tasks

import (
	"regexp"
	"retro-chat-rooms/chat"
	"retro-chat-rooms/discord"
	"time"

	"github.com/bwmarrin/discordgo"
)

func observeRoomMessages(room chat.ChatRoom) {
	if room.DiscordChannel == "" {
		return
	}

	c := chat.RoomMessageEvents[room.ID].Subscribe("discord-bot")
	for message := range c {
		msg := message.(*chat.ChatMessage)
		if !msg.FromDiscord {
			discord.Instance.SendMessage(room.DiscordChannel, msg)
		}
	}
}

func ObserveMessagesToDiscord() {
	rooms := chat.GetAllRooms()

	for _, room := range rooms {
		observeRoomMessages(room)
	}
}

func OnReceiveDiscordMessage(m *discordgo.MessageCreate) {
	content := m.Content

	roomId, found := chat.FindRoomIdByDiscordChannel(m.ChannelID)

	if !found {
		return
	}

	combinedId := chat.GetCombinedId(roomId, m.Author.ID)
	_, found = chat.GetUserByDiscordId(m.Author.ID)
	if !found {
		chat.RegisterUser(chat.ChatUser{
			RoomId:    roomId,
			ID:        combinedId,
			Nickname:  m.Author.Username,
			Color:     chat.USER_COLOR_BLACK,
			DiscordId: m.Author.ID,
			IsAdmin:   false,
		})
	}
	now := time.Now().UTC()

	messageMentionExpr := regexp.MustCompile(`^\s*@([^:]+):\s*(.+)`)

	match := messageMentionExpr.FindStringSubmatch(m.Content)

	var to string = ""

	if len(match) > 1 {
		content = match[2]
		toUser, found := chat.GetUserByNickname(match[1])

		if found {
			to = toUser.ID
		}
	}

	chat.SendMessage(roomId, &chat.ChatMessage{
		Time:            now,
		Message:         content,
		From:            combinedId,
		SpeechMode:      chat.MODE_SAY_TO,
		To:              to,
		IsSystemMessage: false,
		FromDiscord:     true,
	})
}
