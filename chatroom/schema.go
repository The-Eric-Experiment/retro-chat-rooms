package chatroom

import "time"

type PubSubPostMessage struct {
	Room    *Room
	Message *RoomMessage
}

type NicknameColor struct {
	Color string
	Name  string
}

type SpeechMode struct {
	Value string
	Label string
}

type RoomMessage struct {
	Time            time.Time
	Message         string
	From            *RoomUser
	To              *RoomUser
	Privately       bool
	SpeechMode      string
	IsSystemMessage bool
	FromDiscord     bool
}

type RoomUser struct {
	ID                 string
	LastActivity       time.Time
	Nickname           string
	Color              string
	LastPing           time.Time
	LastUserListUpdate time.Time
	SessionIdent       string
	IsDiscordUser      bool
	IsOwner            bool
}
