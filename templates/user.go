package templates

import (
	"html/template"
	"retro-chat-rooms/chat"
	"retro-chat-rooms/routes"
	"strings"
)

func writeNickname(b *strings.Builder, user *chat.ChatUser) {
	b.WriteString("<strong>")

	if user == nil || user.ID == "" {
		b.WriteString("Everyone")
	} else {
		b.WriteString("<font color=\"")
		b.WriteString(user.Color)
		b.WriteString("\">")
		b.WriteString(user.Nickname)
		b.WriteString("</font>")
	}

	b.WriteString("</strong> ")
}

func wrapNicknameWithLink(b *strings.Builder, user *chat.ChatUser) string {
	var buffer strings.Builder

	buffer.WriteString(`<a href="`)
	buffer.WriteString(routes.UrlChatTalk(user.RoomId, user.ID))
	buffer.WriteString(`" target="talk">`)
	buffer.WriteString(b.String())
	buffer.WriteString("</a>")

	return buffer.String()
}

func RenderUsername(userId string, user *chat.ChatUser) template.HTML {
	var buffer strings.Builder

	writeNickname(&buffer, user)

	if userId != user.ID {
		return template.HTML(wrapNicknameWithLink(&buffer, user))
	}

	buffer.WriteString(` <font size="-2">(Me)</font>`)

	return template.HTML(buffer.String())
}
