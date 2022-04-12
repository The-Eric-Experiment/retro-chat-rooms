package chatroom

import "github.com/samber/lo"

func FindRoomByID(id string) *Room {
	room, ok := lo.Find(CHAT_ROOMS, func(r *Room) bool { return r.ID == id })
	if !ok {
		return nil
	}
	return room
}

func FindRoomByDiscordChannel(channel string) *Room {
	room, ok := lo.Find(CHAT_ROOMS, func(r *Room) bool { return r.DiscordChannel == channel })
	if !ok {
		return nil
	}
	return room
}
