package tasks

import (
	"regexp"
	"retro-chat-rooms/chat"
	"retro-chat-rooms/discord"
	"retro-chat-rooms/pubsub"
	"time"

	"github.com/bwmarrin/discordgo"
)

func observeRoomMessages(roomId string, events pubsub.Pubsub) {
	c := events.Subscribe("discord-bot")
	for message := range c {
		switch evt := message.(type) {
		case chat.ChatMessageEvent:
			msg := evt.Message
			if msg != nil && msg.Source != chat.MSG_SOURCE_DISCORD && !msg.IsSystemMessage {
				room, _ := chat.GetSingleRoom(roomId)
				if room.DiscordChannel != "" {
					discord.Instance.SendMessage(room.DiscordChannel, msg)
				}
			}

		}
	}
}

func ObserveMessagesToDiscord() {
	for roomId, events := range chat.RoomEvents {
		go observeRoomMessages(roomId, events)
	}
}

func OnReceiveDiscordMessage(m *discordgo.MessageCreate) {
	content := m.Content

	roomId, found := chat.FindRoomIdByDiscordChannel(m.ChannelID)

	if !found {
		return
	}

	combinedId := chat.GetCombinedId(roomId, m.Author.ID)
	user, found := chat.GetUserByDiscordId(m.Author.ID)
	if !found {
		user = chat.ChatUser{
			RoomId:    roomId,
			ID:        combinedId,
			Nickname:  m.Author.Username,
			Color:     chat.USER_COLOR_BLACK,
			DiscordId: m.Author.ID,
			IsAdmin:   false,
			Client: chat.ClientInfo{
				Plat: chat.CLIENT_PLATFORM_DISCORD,
			},
		}
		chat.RegisterUser(user)
	}
	now := time.Now().UTC()

	messageMentionExpr := regexp.MustCompile(`^\s*@([^:]+):\s*(.+)`)

	match := messageMentionExpr.FindStringSubmatch(m.Content)

	var to string = ""
	involvedUsers := []chat.ChatUser{user}

	if len(match) > 1 {
		content = match[2]
		toUser, found := chat.GetUserByNickname(match[1])

		if found {
			to = toUser.ID
			involvedUsers = append(involvedUsers, toUser)
		}
	}

	chat.SendMessage(&chat.ChatMessage{
		RoomID:          roomId,
		Time:            now,
		Message:         content,
		From:            user.ID,
		SpeechMode:      chat.MODE_SAY_TO,
		To:              to,
		IsSystemMessage: false,
		Source:          chat.ClientInfoToMsgSource(user.Client),
		ShowClientIcon:  true,
		InvolvedUsers:   involvedUsers,
	})
}
