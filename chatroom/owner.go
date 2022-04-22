package chatroom

import (
	"retro-chat-rooms/config"
)

func LoginOwner(userId string, config config.OwnerChatUserConfig) {
	// if OwnerRoomUser == nil {
	// 	OwnerRoomUser = &RoomUser{
	// 		ID:                 config.Id,
	// 		LastActivity:       time.Now().UTC(),
	// 		Nickname:           config.Nickname,
	// 		Color:              config.Color,
	// 		LastPing:           time.Now().UTC(),
	// 		LastUserListUpdate: time.Now().UTC(),
	// 		IsDiscordUser:      config.DiscordId != "",
	// 		IsOwner:            true,
	// 	}
	// }
}

func LogoutOwner() {
	// OwnerRoomUser.SessionIdent = ""
	// if !OwnerRoomUser.IsDiscordUser {
	// 	OwnerRoomUser = nil
	// }
}
