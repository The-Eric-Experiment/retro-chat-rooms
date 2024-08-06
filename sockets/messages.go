package sockets

import (
	"fmt"
	"regexp"
	"retro-chat-rooms/chat"
	"retro-chat-rooms/helpers"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

func writeNickname(b *strings.Builder, user *chat.ChatUser) {
	b.WriteString("[B]")

	if user == nil || user.ID == "" {
		b.WriteString("Everyone")
	} else {
		b.WriteString("[C=")
		b.WriteString(HtmlToDelphiColor(user.Color))
		b.WriteString("]")
		b.WriteString(user.Nickname)
		b.WriteString("[/C]")
	}

	b.WriteString("[/B] ")
}

func writeMessage(b *strings.Builder, msg chat.ChatMessage) {
	if !msg.IsSystemMessage {
		b.WriteString(msg.Message)
		return
	}

	b.WriteString(
		strings.ReplaceAll(msg.Message, "{nickname}", "[C="+HtmlToDelphiColor(msg.SystemMessageSubject.Color)+"]"+msg.SystemMessageSubject.Nickname+"[/C]"),
	)
}

func writeUserMessageDescriptor(b *strings.Builder, speechMode string, message chat.ChatMessage, messageRendered func()) {
	writeNickname(b, message.GetFrom())
	if message.Privately {
		b.WriteString("privately ")
	}
	b.WriteString("[I]" + speechMode + "[/I] ")
	writeNickname(b, message.GetTo())
	b.WriteString(": ")
	messageRendered()
}

func RenderClientMessage(message chat.ChatMessage) string {
	var buffer strings.Builder

	buffer.WriteString("[C=" + HtmlToDelphiColor("#DDDDDD") + "]")
	buffer.WriteString("\\[")
	buffer.WriteString(helpers.FormatTimestamp24H(message.Time))
	buffer.WriteString("\\] ")
	buffer.WriteString("[/C]")

	if !message.IsSystemMessage {
		switch message.SpeechMode {
		case "says-to":
			writeUserMessageDescriptor(&buffer, "says to", message, func() {
				writeMessage(&buffer, message)
			})
		case "screams-at":
			writeUserMessageDescriptor(&buffer, "screams at", message, func() {
				buffer.WriteString("[B]")
				writeMessage(&buffer, message)
				buffer.WriteString("[/B]")
			})
		case "whispers-to":
			writeUserMessageDescriptor(&buffer, "whispers to", message, func() {
				buffer.WriteString("[I]")
				writeMessage(&buffer, message)
				buffer.WriteString("[/I]")
			})
		}
	} else {
		writeMessage(&buffer, message)
	}

	return buffer.String()
}

func PushUserJoined(conn ISocket, user chat.ChatUser) {
	response := SerializeMessage(SERVER_USER_LIST_ADD, &ServerUserListAdd{
		UserID:   user.ID,
		Nickname: user.Nickname,
		Color:    HtmlToDelphiColor(user.Color),
		RoomID:   user.RoomId,
	})
	conn.Write(response)
}

func PushUserLeft(conn ISocket, user chat.ChatUser) {
	response := SerializeMessage(SERVER_USER_LIST_REMOVE, &ServerUserListRemove{
		UserID: user.ID,
		RoomID: user.RoomId,
	})
	conn.Write(response)
}

func PushMessage(conn ISocket, msg chat.ChatMessage) {
	from := msg.GetFrom()
	to := msg.GetTo()

	connUser := conn.GetUser()

	if to != nil && to.ID != connUser.ID {
		return
	}

	toID := ""
	toNickname := "everyone"

	if to != nil {
		toID = to.ID
		toNickname = to.Nickname
	}

	Text := RenderClientMessage(msg)

	response := SerializeMessage(SERVER_MESSAGE_SENT, &ServerMessageSent{
		FromNickname: from.Nickname,
		FromID:       from.ID,
		ToID:         toID,
		ToNickname:   toNickname,
		Text:         Text,
	})
	conn.Write(response)
}

func registerUser(conn ISocket, msg string) {
	content := DeserializeMessage(RegisterUser{}, msg)

	userId := chat.GetCombinedId(content.RoomID, uuid.NewString())
	color := DelphiToHtmlColor(content.Color)

	_, err := chat.RegisterUser(chat.ChatUser{
		ID:        userId,
		Nickname:  content.Nickname,
		Color:     color,
		RoomId:    content.RoomID,
		DiscordId: "",
		IsAdmin:   false,
		IsWebUser: false,
	})

	if err != nil {
		errMsg := err.Error()
		response := SerializeMessage(SERVER_ERROR, &ServerError{Message: errMsg})
		conn.Write(response)
		return
	}

	users := chat.GetRoomUsers(content.RoomID)

	for _, user := range users {
		userMsg := SerializeMessage(SERVER_USER_LIST_ADD, &ServerUserListAdd{
			UserID:   user.ID,
			Nickname: user.Nickname,
			Color:    HtmlToDelphiColor(user.Color),
			RoomID:   user.RoomId,
		})

		conn.Write(userMsg)
	}

	user, ok := chat.GetUser(userId)

	if !ok {
		response := SerializeMessage(SERVER_ERROR, &ServerError{Message: "User not found!"})
		conn.Write(response)
		return
	}

	response := SerializeMessage(SERVER_USER_REGISTRATION_SUCCESS, &RegisterUserResponse{UserID: userId, Nickname: user.Nickname, Color: user.Color, RoomID: user.RoomId})

	InternalEvents.Publish(InternalUserRegisteredEvent{ConnectionID: conn.ID(), User: user})

	conn.Write(response)
}

func sendMessage(conn ISocket, msg string) {
	content := DeserializeMessage(SendMessage{}, msg)

	user, foundUser := chat.GetUser(content.UserID)

	if !foundUser {
		return
	}

	involvedUsers := []chat.ChatUser{user}

	toUser, foundToUser := chat.GetUser(content.To)

	if foundToUser {
		involvedUsers = append(involvedUsers, toUser)
	}

	privately, err := strconv.ParseBool(content.Privately)

	if err != nil {
		//TODO: DEAL WITH THIS
	}

	chat.SendMessage(content.RoomID, &chat.ChatMessage{
		Time:                 time.Now().UTC(),
		Message:              content.Message,
		From:                 content.UserID,
		To:                   content.To,
		Privately:            privately,
		SpeechMode:           content.SpeechMode,
		IsSystemMessage:      false,
		FromDiscord:          false,
		SystemMessageSubject: user,
		InvolvedUsers:        involvedUsers,
	})

}

func ping(conn ISocket, msg string) {
	content := DeserializeMessage(Ping{}, msg)
	fmt.Println("Acknowledged", content.UserId)
}

func colorListRequest(conn ISocket, msg string) {
	fmt.Println("Color list request")
	colors := make([]string, len(chat.NICKNAME_COLORS))
	names := make([]string, len(chat.NICKNAME_COLORS))

	for idx, def := range chat.NICKNAME_COLORS {
		colors[idx] = HtmlToDelphiColor(def.Color)
		names[idx] = def.Name
	}

	response := SerializeMessage(SERVER_COLOR_LIST, &ColorList{Colors: colors, ColorNames: names})

	conn.Write(response)
}

func roomListRequest(conn ISocket, msg string) {
	fmt.Println("Room list request")
	rooms := chat.GetAllRooms()
	conn.Write(SerializeMessage(SERVER_ROOM_LIST_START, &RoomListStart{}))
	for _, room := range rooms {
		conn.Write(SerializeMessage(SERVER_ROOM_LIST_ITEM, &RoomListItem{
			RoomId: room.ID,
			Name:   room.Name,
		}))
	}
	conn.Write(SerializeMessage(SERVER_ROOM_LIST_END, &RoomListEnd{}))
}

func HtmlToDelphiColor(color string) string {
	r, _ := regexp.Compile(`#(?P<R>.{2})(?P<G>.{2})(?P<B>.{2})`)
	rIdx := r.SubexpIndex("R")
	gIdx := r.SubexpIndex("G")
	bIdx := r.SubexpIndex("B")
	match := r.FindStringSubmatch(color)

	return strings.Join([]string{"$", match[bIdx], match[gIdx], match[rIdx]}, "")
}

func DelphiToHtmlColor(color string) string {
	r, _ := regexp.Compile(`\$(?P<B>.{2})(?P<G>.{2})(?P<R>.{2})`)
	rIdx := r.SubexpIndex("R")
	gIdx := r.SubexpIndex("G")
	bIdx := r.SubexpIndex("B")
	match := r.FindStringSubmatch(color)

	return strings.Join([]string{"#", match[rIdx], match[gIdx], match[bIdx]}, "")
}

func ProcessMessage(conn ISocket, message []byte) {
	msg := string(message)
	t, msgContent := determineMessageType(msg)
	switch t {
	case CLIENT_REGISTER_USER:
		registerUser(conn, msgContent)
	case CLIENT_COLOR_LIST_REQUEST:
		colorListRequest(conn, msgContent)
	case CLIENT_ROOM_LIST_REQUEST:
		roomListRequest(conn, msgContent)
	case CLIENT_PING:
		ping(conn, msgContent)
	case CLIENT_SEND_MESSAGE:
		sendMessage(conn, msgContent)
	default:
		conn.Write(`0 "Invalid client message"`)
	}
}
