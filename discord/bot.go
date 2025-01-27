package discord

import (
	"fmt" //to print errors

	"retro-chat-rooms/chat"
	"retro-chat-rooms/config"
	"strings"

	"github.com/bwmarrin/discordgo" //discordgo package from the repo of bwmarrin .
)

func formatMessageForDiscord(m *chat.ChatMessage) discordgo.WebhookParams {
	from, _ := chat.GetUser(m.From)

	if m.IsSystemMessage {
		return discordgo.WebhookParams{
			Username: "System",
			Content:  strings.Replace(m.Message, "{nickname}", from.Nickname, -1),
		}
	}

	message := ""

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

	return discordgo.WebhookParams{
		Content:  message,
		Username: strings.Replace("{nickname} @ Old'aVista Chat!", "{nickname}", from.Nickname, -1),
		AllowedMentions: &discordgo.MessageAllowedMentions{
			Parse: []discordgo.AllowedMentionType{
				"users",
			},
		},
	}
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

	params := formatMessageForDiscord(message)

	_, err := bot.session.WebhookExecute(
		config.Current.DiscordWebhookId,
		config.Current.DiscordWebhookToken,
		true,
		&params,
	)

	if err != nil {
		fmt.Printf("There was an error sending discord message to %s: %s\n", channel, err.Error())
	}
}

func (bot *DiscordBot) OnReceiveMessage(fn func(m *discordgo.MessageCreate)) {
	if bot.session == nil {
		return
	}

	bot.session.AddHandler(messageListen(fn))
}

func messageListen(fn func(m *discordgo.MessageCreate)) func(s *discordgo.Session, m *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		// Ignore all messages created by the bot itself or by the webhook
		if m.Author.ID == s.State.User.ID || (m.Author.ID == m.WebhookID && m.WebhookID == config.Current.DiscordWebhookId) {
			return
		}

		fn(m)
	}
}

var Instance = DiscordBot{}
