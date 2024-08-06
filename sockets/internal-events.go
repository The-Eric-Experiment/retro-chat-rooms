package sockets

import (
	"retro-chat-rooms/chat"
	"retro-chat-rooms/pubsub"
)

type InternalUserRegisteredEvent struct {
	ConnectionID string
	User         chat.ChatUser
}

var InternalEvents = pubsub.NewPubsub()
