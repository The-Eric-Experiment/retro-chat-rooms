package templates

import (
	"html/template"
	chatroom "retro-chat-rooms/chatroom"
	"strings"
	"time"
)

func formatMessage(msg *chatroom.RoomMessage) template.HTML {
	if !msg.IsSystemMessage {
		return template.HTML(template.HTMLEscapeString(msg.Message))
	}
	transformed := strings.ReplaceAll(msg.Message, "{nickname}", "<strong><font color=\""+msg.From.Color+"\">"+msg.From.Nickname+"</font></strong>")
	return template.HTML(transformed)
}

func formatNickname(user *chatroom.RoomUser) template.HTML {
	if user == nil {
		return template.HTML("<strong>Everyone</strong>")
	}

	return template.HTML("<strong><font color=\"" + user.Color + "\">" + user.Nickname + "</font></strong>")
}

func isUserNotNil(input *chatroom.RoomUser) bool {
	return input != nil
}

func formatTime(t time.Time) string {
	return t.Format("03:04:05 PM")
}

func isMessageVisible(userId string, message *chatroom.RoomMessage) bool {
	if !message.Privately || message.To == nil {
		return true
	}

	return userId == message.To.ID || userId == message.From.ID
}

func isMessageToSelf(userId string, message *chatroom.RoomMessage) bool {
	if message.To == nil {
		return false
	}

	return userId == message.To.ID
}

func countUsers(input []*chatroom.RoomUser) int {
	return len(input)
}

func hasStrings(input []string) bool {
	return len(input) > 0
}
