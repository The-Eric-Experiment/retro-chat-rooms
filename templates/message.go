package templates

import (
	"html/template"
	"retro-chat-rooms/chat"
	"retro-chat-rooms/helpers"
	"strings"
)

func messageToUserWrap(b *strings.Builder) string {
	var buffer strings.Builder
	buffer.WriteString(`<table bgcolor="#E0E0E0" border="1" BORDERCOLOR="#DDDDDD" width="100%" cellspacing="0"cellpadding="2"><tr><td>`)
	buffer.WriteString(b.String())
	buffer.WriteString("</td></tr></table>")
	return buffer.String()
}

func writeMessage(b *strings.Builder, msg *chat.ChatMessage) {
	if !msg.IsSystemMessage {
		b.WriteString(template.HTMLEscapeString(msg.Message))
		return
	}

	b.WriteString(
		strings.ReplaceAll(msg.Message, "{nickname}", "<strong><font color=\""+msg.SystemMessageSubject.Color+"\">"+msg.SystemMessageSubject.Nickname+"</font></strong>"),
	)
}

func writeUserMessageDescriptor(b *strings.Builder, speechMode string, message *chat.ChatMessage, messageRendered func()) {
	from := message.GetFrom()
	to := message.GetTo()

	writeNickname(b, from)
	if message.Privately {
		b.WriteString("<i>privately</i> ")
	}
	b.WriteString(speechMode)
	b.WriteString(" ")
	writeNickname(b, to)
	b.WriteString(": ")
	messageRendered()
}

func RenderMessage(userId string, message *chat.ChatMessage) template.HTML {
	var buffer strings.Builder

	buffer.WriteString("[")
	buffer.WriteString(helpers.FormatTimestamp(message.Time))
	buffer.WriteString("] ")

	if !message.IsSystemMessage {
		switch message.SpeechMode {
		case chat.MODE_SAY_TO:
			writeUserMessageDescriptor(&buffer, "says to", message, func() {
				writeMessage(&buffer, message)
			})
		case chat.MODE_SCREAM_AT:
			writeUserMessageDescriptor(&buffer, "screams at", message, func() {
				buffer.WriteString(`<font size="4"><strong>`)
				writeMessage(&buffer, message)
				buffer.WriteString("</strong></font>")
			})
		case chat.MODE_WHISPER_TO:
			writeUserMessageDescriptor(&buffer, "whispers to", message, func() {
				buffer.WriteString(`<font size="2"><i>`)
				writeMessage(&buffer, message)
				buffer.WriteString("</i></font>")
			})
		}
	} else {
		writeMessage(&buffer, message)
	}

	if message.To == userId {
		return template.HTML(messageToUserWrap(&buffer))
	}

	buffer.WriteString("<br>")

	return template.HTML(buffer.String())

}
