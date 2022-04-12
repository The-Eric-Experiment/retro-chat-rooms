package tasks

import (
	"retro-chat-rooms/chatroom"
	"retro-chat-rooms/discord"
	"retro-chat-rooms/pubsub"
)

func ObserveMessagesForDiscord() {
	c := pubsub.Messaging.Subscribe("message-sent")
	for message := range c {
		msg := message.(*chatroom.PubSubPostMessage)
		if !msg.Message.FromDiscord {
			discord.Instance.SendMessage(msg.Room.DiscordChannel, msg.Message)
		}
	}
}
