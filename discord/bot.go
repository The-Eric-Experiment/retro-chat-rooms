package discord

import (
	"fmt" //to print errors

	"retro-chat-rooms/chatroom"
	"retro-chat-rooms/config"
	"strings"

	"github.com/bwmarrin/discordgo" //discordgo package from the repo of bwmarrin .
)

func formatMessageForDiscord(m *chatroom.RoomMessage) string {
	message := "**" + m.From.Nickname + "** "

	if m.IsSystemMessage {
		return strings.Replace(m.Message, "{nickname}", message, -1)
	}

	message += "*" + modeTransform[m.SpeechMode] + "* "

	if m.To != nil {
		// Yuck
		if m.To.IsAdmin && config.Current.OwnerChatUser.DiscordId != "" {
			message += "<@" + config.Current.OwnerChatUser.DiscordId + ">"
		} else if m.To.IsDiscordUser {
			message += "<@" + m.To.ID + ">"
		} else {
			message += "**" + m.To.Nickname + "**"
		}

	} else {
		message += "**Everyone**"
	}

	message += ": "

	switch m.SpeechMode {
	case chatroom.MODE_SAY_TO:
		message += m.Message
	case chatroom.MODE_SCREAM_AT:
		message += "**" + strings.ToUpper(m.Message) + "**"
	case chatroom.MODE_WHISPER_TO:
		message += "*" + strings.ToLower(m.Message) + "*"
	}

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

func (bot *DiscordBot) SendMessage(channel string, message *chatroom.RoomMessage) {
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

// func StartDiscordBot() {

// 	// Wait here until CTRL-C or other term signal is received.
// 	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
// 	sc := make(chan os.Signal, 1)
// 	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
// 	<-sc

// 	// Cleanly close down the Discord session.
// 	dg.Close()
// }

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
