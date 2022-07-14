package discord

import (
	"fmt" //to print errors

	"retro-chat-rooms/chat"
	"retro-chat-rooms/config"
	"retro-chat-rooms/helpers"
	"strings"

	"github.com/bwmarrin/discordgo" //discordgo package from the repo of bwmarrin .
)

func formatMessageForDiscord(m *chat.ChatMessage) string {
	from, _ := chat.GetUser(m.From)
	message := " \n**" + from.Nickname + "** *(" + helpers.FormatTimestamp(m.Time) + ")*:  \n"

	if m.IsSystemMessage {
		message = "**" + m.SystemMessageSubject.Nickname + "** "
		return strings.Replace(m.Message, "{nickname}", message, -1)
	}

	// message += "*" + modeTransform[m.SpeechMode] + "* "

	if m.To != "" {
		to, _ := chat.GetUser(m.To)
		// Yuck
		if to.IsAdmin && config.Current.OwnerChatUser.DiscordId != "" {
			message += "<@" + config.Current.OwnerChatUser.DiscordId + "> "
		} else if to.DiscordId != "" {
			message += "<@" + to.DiscordId + "> "
		} else {
			message += "**" + to.Nickname + "** "
		}
	}

	switch m.SpeechMode {
	case chat.MODE_SAY_TO:
		message += m.Message
	case chat.MODE_SCREAM_AT:
		message += "**" + strings.ToUpper(m.Message) + "**"
	case chat.MODE_WHISPER_TO:
		message += "*" + strings.ToLower(m.Message) + "*"
	}

	message += "\n"

	return message
}

type DiscordBot struct {
	session *discordgo.Session
}

func (bot *DiscordBot) Connect() {
	if config.Current.DiscordBotToken == "" {
		return
	}

	dg, err := discordgo.New("Bot " + config.Current.DiscordBotToken)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	dg.Identify.Intents = discordgo.IntentsGuildMessages

	err = dg.Open()

	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	bot.session = dg
}

func (bot *DiscordBot) Close() {
	if bot.session == nil {
		return
	}
	bot.session.Close()
}

func (bot *DiscordBot) SendMessage(channel string, message *chat.ChatMessage) {
	if bot.session == nil || channel == "" || message.Privately {
		return
	}

	text := formatMessageForDiscord(message)

	bot.session.ChannelMessageSend(channel, text)
}

func (bot *DiscordBot) OnReceiveMessage(fn func(m *discordgo.MessageCreate)) {
	if bot.session == nil {
		return
	}

	bot.session.AddHandler(messageListen(fn))
}

func messageListen(fn func(m *discordgo.MessageCreate)) func(s *discordgo.Session, m *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		// Ignore all messages created by the bot itself
		// This isn't required in this specific example but it's a good practice.
		if m.Author.ID == s.State.User.ID {
			return
		}

		fn(m)
	}
}

var Instance = DiscordBot{}
