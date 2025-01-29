package chat

import (
	"time"

	"github.com/samber/lo"
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
	RoomID               string
	Time                 time.Time
	Message              string
	From                 string
	To                   string
	Privately            bool
	SpeechMode           string
	IsSystemMessage      bool
	SystemMessageSubject *ChatUser
	FromDiscord          bool
	InvolvedUsers        []ChatUser
}

func (m *ChatMessage) GetFrom() *ChatUser {
	if m.From == "" {
		return nil
	}

	user, found := lo.Find(m.InvolvedUsers, func(u ChatUser) bool {
		return u.ID == m.From
	})

	if !found {
		return nil
	}

	return &user
}

func (m *ChatMessage) GetTo() *ChatUser {
	if m.To == "" {
		return nil
	}

	user, found := lo.Find(m.InvolvedUsers, func(u ChatUser) bool {
		return u.ID == m.To
	})

	if !found {
		return nil
	}

	return &user
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

type ChatUserKickedEvent struct {
	UserID  string
	Message string
}
