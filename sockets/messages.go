package sockets

import (
	"fmt"
	"reflect"
	"retro-chat-rooms/chat"
	"retro-chat-rooms/floodcontrol"
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

func PushUserKickedMessage(conn ISocket, evt chat.ChatUserKickedEvent) {
	if evt.UserID != conn.GetUser().ID {
		return
	}
	chat.DeregisterUser(conn.GetUser().ID)
	response := SerializeMessage(SERVER_USER_KICKED, &ServerUserKicked{
		Reason: evt.Message,
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

func PushMessage(conn ISocket, msg chat.ChatMessage, isHistory bool) {
	socketUserState := NewSocketsUserState(conn)

	if floodcontrol.IsIPBanned(socketUserState.GetUserIP()) {
		return
	}

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
		RoomID:               msg.RoomID,
		From:                 SerializeSubObject(fromUser),
		To:                   SerializeSubObject(toUser),
		Privately:            strconv.FormatBool(msg.Privately),
		SpeechMode:           msg.SpeechMode,
		Time:                 helpers.FormatTimestamp24H(msg.Time),
		IsSystemMessage:      strconv.FormatBool(msg.IsSystemMessage),
		SystemMessageSubject: SerializeSubObject(systemMessageSubjectUser),
		Message:              msg.Message,
		IsHistory:            strconv.FormatBool(isHistory),
	}

	response := SerializeMessage(SERVER_MESSAGE_SENT, &message)

	conn.Write(response)
}

func registerUser(conn ISocket, msg string) {
	content := DeserializeMessage(RegisterUser{}, msg)

	socketUserState := NewSocketsUserState(conn)

	userId := chat.GetCombinedId(content.RoomID, uuid.NewString())
	color := content.Color

	errors := make([]string, 0)

	newUser := chat.ChatUser{
		ID:        userId,
		Nickname:  content.Nickname,
		Color:     color,
		RoomId:    content.RoomID,
		DiscordId: "",
		IsAdmin:   false,
		IsWebUser: false,
	}

	chat.ValidateUser(&socketUserState, newUser, &errors)

	// YUCK THESE IFS
	if len(errors) == 0 {
		_, err := chat.RegisterUser(newUser)
		if err != nil && err.Error() == "user exists" {
			errors = append(errors, "Someone is already using this Nickname, try a different one.")
		} else if err != nil {
			errors = append(errors, "Couldn't register user, try again.")
		}
	}

	if len(errors) > 0 {
		response := SerializeMessage(SERVER_ERROR, &ServerError{Message: errors[0]})
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
				PushMessage(conn, *message, true)
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
		return
	}

	socketUserState := NewSocketsUserState(conn)

	message, isValid := chat.ValidateMessage(&socketUserState, chat.ChatMessage{
		RoomID:               content.RoomID,
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

	if !isValid {
		return
	}

	if floodcontrol.IsIPBanned(socketUserState.GetUserIP()) {
		go chat.RoomEvents[content.RoomID].Publish(chat.ChatUserKickedEvent{
			UserID:  user.ID,
			Message: "You have been temporarily kicked out for flooding.",
		})
	}

	chat.SendMessage(&message)
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
	roomIds := make([]string, len(rooms))
	names := make([]string, len(rooms))

	for idx, def := range rooms {
		roomIds[idx] = def.ID
		names[idx] = def.Name
	}

	response := SerializeMessage(SERVER_ROOM_LIST, &RoomList{RoomIDs: roomIds, RoomNames: names})

	conn.Write(response)
}

func ProcessMessage(conn ISocket, message []byte) {
	msgType, _, msgContent, err := determineMessageType(message)
	if err != nil {
		return
	}
	switch msgType {
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
	}
}
