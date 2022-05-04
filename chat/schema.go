package chat

import (
	"time"
)

type PubSubPostMessage struct {
	RoomId  string
	Message *ChatMessage
}

type NicknameColor struct {
	Color string
	Name  string
}

type SpeechMode struct {
	Value string
	Label string
}

type ChatRoom struct {
	ID                 string
	Name               string
	Description        string
	Color              string
	DiscordChannel     string
	LastUserListUpdate time.Time
	TextColor          string
}

type ChatMessage struct {
	Time                 time.Time
	Message              string
	From                 string
	To                   string
	Privately            bool
	SpeechMode           string
	IsSystemMessage      bool
	SystemMessageSubject ChatUser
	FromDiscord          bool
}

type ChatUser struct {
	ID        string
	Nickname  string
	Color     string
	DiscordId string
	IsAdmin   bool
	RoomId    string
}

func (user ChatUser) IsDiscordUser() bool {
	return user.DiscordId != ""
}

type UserListUpdatedEvent struct {
	RoomId string
}

type ChatEvent struct {
	Message          *ChatMessage
	IsUserListUpdate bool
}
