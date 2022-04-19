package chatroom

import (
	"retro-chat-rooms/config"
	"time"
)

var OwnerRoomUser *RoomUser

func InitializeOwner(config config.OwnerChatUserConfig) {
	if config.DiscordId == "" {
		return
	}

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

func LoginOwner(config config.OwnerChatUserConfig, ident string) {
	if OwnerRoomUser == nil {
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
	OwnerRoomUser.SessionIdent = ident
}

func LogoutOwner() {
	OwnerRoomUser.SessionIdent = ""
	if !OwnerRoomUser.IsDiscordUser {
		OwnerRoomUser = nil
	}
}
