package templates

import (
	"retro-chat-rooms/chat"
	"retro-chat-rooms/helpers"
	"time"
)

func formatTime(t time.Time) string {
	return helpers.FormatTimestamp(t)
}

func countUsers(roomId string) int {
	return len(chat.GetRoomOnlineUsers(roomId))
}

func hasStrings(input []string) bool {
	return len(input) > 0
}
