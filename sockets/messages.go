package sockets

import (
	"fmt"
	"reflect"
	"retro-chat-rooms/chat"
	"retro-chat-rooms/helpers"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

func PushUserJoined(conn ISocket, user chat.ChatUser) {
	response := SerializeMessage(SERVER_USER_LIST_ADD, &ServerUserListAdd{
		UserID:   user.ID,
		Nickname: user.Nickname,
		Color:    user.Color,
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

func SerializeSubObject(val interface{}) string {
	if val == nil || (val != nil && reflect.ValueOf(val).IsNil()) {
		return ""
	}

	serialized := SerializeObject(val)

	return strings.ReplaceAll(strings.ReplaceAll(serialized, "\\", "\\\\"), "\"", "\\\"")
}

func PushMessage(conn ISocket, msg chat.ChatMessage) {
	connUser := conn.GetUser()

	if msg.Privately && msg.To != "" && (msg.To != connUser.ID && msg.From != connUser.ID) {
		return
	}

	from := msg.GetFrom()
	to := msg.GetTo()

	var fromUser *ServerUserListAdd = nil
	var toUser *ServerUserListAdd = nil
	var systemMessageSubjectUser *ServerUserListAdd = nil

	if from != nil {
		fromUser = &ServerUserListAdd{
			UserID:   from.ID,
			Nickname: from.Nickname,
			Color:    from.Color,
			RoomID:   from.RoomId,
		}
	}

	if to != nil {
		toUser = &ServerUserListAdd{
			UserID:   to.ID,
			Nickname: to.Nickname,
			Color:    to.Color,
			RoomID:   to.RoomId,
		}
	}

	if msg.SystemMessageSubject != nil {
		systemMessageSubjectUser = &ServerUserListAdd{
			UserID:   msg.SystemMessageSubject.ID,
			Nickname: msg.SystemMessageSubject.Nickname,
			Color:    msg.SystemMessageSubject.Color,
			RoomID:   msg.SystemMessageSubject.RoomId,
		}
	}

	message := ServerMessageSent{
		RoomID:               from.RoomId,
		From:                 SerializeSubObject(fromUser),
		To:                   SerializeSubObject(toUser),
		Privately:            strconv.FormatBool(msg.Privately),
		SpeechMode:           msg.SpeechMode,
		Time:                 helpers.FormatTimestamp24H(msg.Time),
		IsSystemMessage:      strconv.FormatBool(msg.IsSystemMessage),
		SystemMessageSubject: SerializeSubObject(systemMessageSubjectUser),
		Message:              msg.Message,
	}

	response := SerializeMessage(SERVER_MESSAGE_SENT, &message)

	conn.Write(response)
}

func registerUser(conn ISocket, msg string) {
	content := DeserializeMessage(RegisterUser{}, msg)

	userId := chat.GetCombinedId(content.RoomID, uuid.NewString())
	color := content.Color

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
			Color:    user.Color,
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

	go func() {
		messageHistory, hasMessages := chat.GetUserMessageList(userId)

		if hasMessages {
			for _, message := range messageHistory {
				PushMessage(conn, *message)
			}
		}
	}()
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
		SystemMessageSubject: &user,
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
		colors[idx] = def.Color
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
