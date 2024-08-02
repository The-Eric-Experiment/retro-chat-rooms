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
	IntroMessage       string
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
	InvolvedUsers        []ChatUser
}

type ChatUser struct {
	ID        string
	Nickname  string
	Color     string
	DiscordId string
	IsAdmin   bool
	RoomId    string
	IsWebUser bool
}

func (user ChatUser) IsDiscordUser() bool {
	return user.DiscordId != ""
}

type ChatMessageEvent struct {
	Message *ChatMessage
}

type ChatUserJoinedEvent struct {
	User ChatUser
}

type ChatUserLeftEvent struct {
	User ChatUser
}
