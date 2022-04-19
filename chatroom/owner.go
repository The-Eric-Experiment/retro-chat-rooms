package chatroom

import (
	"retro-chat-rooms/config"
	"time"
)

var OwnerRoomUser *RoomUser

func InitializeOwner(config config.OwnerChatUserConfig) {
	OwnerRoomUser = &RoomUser{
		ID:                 config.Id,
		LastActivity:       time.Now().UTC(),
		Nickname:           config.Nickname,
		Color:              config.Color,
		LastPing:           time.Now().UTC(),
		LastUserListUpdate: time.Now().UTC(),
		SessionIdent:       "",
		IsDiscordUser:      config.DiscordId != "",
		IsOwner:            true,
	}
}

func LoginOwner(ident string) {
	OwnerRoomUser.SessionIdent = ident
}
