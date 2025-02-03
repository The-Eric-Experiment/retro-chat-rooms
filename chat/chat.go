package chat

import (
	"errors"
	"html/template"
	"retro-chat-rooms/config"
	"retro-chat-rooms/helpers"
	"retro-chat-rooms/pubsub"
	"strings"
	"sync"
	"time"

	"github.com/lucasb-eyer/go-colorful"
	"github.com/samber/lo"
)

var (
	// List of rooms available
	rooms map[string]ChatRoom = make(map[string]ChatRoom)

	// This is the ordered list of room ids,
	// used to return the room list in order.
	roomKeys []string = make([]string, 0)

	// Key/Value list of all users
	users map[string]ChatUser = make(map[string]ChatUser)

	// Key/Value list of Chat messages per user
	userMessages map[string][]*ChatMessage = make(map[string][]*ChatMessage)

	roomMessageHistory map[string][]*ChatMessage = make(map[string][]*ChatMessage)

	// key/value list of user ids per room
	roomUsers map[string][]string = make(map[string][]string)

	userLastUserListChange map[string]time.Time = make(map[string]time.Time)
	roomLastUserListChange map[string]time.Time = make(map[string]time.Time)
	userPings              map[string]time.Time = make(map[string]time.Time)

	RoomEvents map[string]pubsub.Pubsub = make(map[string]pubsub.Pubsub)

	mutex = sync.Mutex{}
)

func determineTextColor(color string) string {
	c, err := colorful.Hex(color)
	if err != nil {
		return "#000000"
	}

	_, _, luminance := c.HSLuv()

	if (luminance * 100) > 60 {
		return "#000000"
	}

	return "#FFFFFF"
}

func userListUpdated(roomId string, event interface{}) {
	defer mutex.Unlock()
	mutex.Lock()
	roomLastUserListChange[roomId] = helpers.Now().UTC()
	RoomEvents[roomId].Publish(event)
}

func instantiateUser(user ChatUser) error {
	defer mutex.Unlock()
	mutex.Lock()
	now := time.Now().UTC()

	// Prevent XSS
	user.Nickname = template.HTMLEscapeString(user.Nickname)

	_, found := users[user.ID]

	if found {
		return errors.New("user exists")
	}

	comparisonNickname := strings.Trim(strings.ToLower(user.Nickname), " ")
	for _, u := range users {
		if strings.Trim(strings.ToLower(u.Nickname), " ") == comparisonNickname {
			return errors.New("nickname selected")
		}
	}

	users[user.ID] = user

	userMessages[user.ID] = make([]*ChatMessage, 0)
	roomUsers[user.RoomId] = append(roomUsers[user.RoomId], user.ID)
	userLastUserListChange[user.ID] = now
	userPings[user.ID] = now

	return nil
}

func removeUser(combinedId string) ChatUser {
	defer mutex.Unlock()
	mutex.Lock()
	user, found := users[combinedId]
	if !found || (user.IsAdmin && user.IsDiscordUser()) {
		return ChatUser{}
	}

	roomUsers[user.RoomId] = lo.Filter(roomUsers[user.RoomId],
		func(uid string, _ int) bool {
			return uid != combinedId
		})

	// Just making sure instances are gone
	for i := range userMessages[combinedId] {
		userMessages[combinedId][i] = nil
	}

	delete(userMessages, combinedId)
	delete(users, combinedId)
	delete(userLastUserListChange, combinedId)
	delete(userPings, combinedId)

	return user
}

func GetCombinedId(roomId string, userId string) string {
	return helpers.GenerateUniqueID(roomId + userId)
}

func RegisterAdmin(roomId string) string {
	cfg := config.Current.OwnerChatUser

	combinedId := helpers.GenerateUniqueID(roomId + cfg.Id)

	_, found := users[combinedId]

	if found {
		return combinedId
	}

	RegisterUser(ChatUser{
		ID:        combinedId,
		Nickname:  cfg.Nickname,
		Color:     cfg.Color,
		DiscordId: cfg.DiscordId,
		IsAdmin:   true,
		RoomId:    roomId,
		IsWebUser: false,
		Client:    "",
	})

	return combinedId
}

func InitializeRooms() {
	for _, cr := range config.Current.Rooms {
		room := ChatRoom{
			ID:                 cr.ID,
			Name:               cr.Name,
			Description:        cr.Description,
			Color:              cr.Color,
			DiscordChannel:     cr.DiscordChannel,
			LastUserListUpdate: time.Now().UTC(),
			IntroMessage:       cr.ChatRoomIntroMessage,
			TextColor:          determineTextColor(cr.Color),
		}
		roomKeys = append(roomKeys, cr.ID)
		rooms[room.ID] = room
		roomUsers[room.ID] = make([]string, 0)
		roomMessageHistory[room.ID] = make([]*ChatMessage, 0)
		roomLastUserListChange[room.ID] = time.Now().UTC()
		RoomEvents[room.ID] = pubsub.NewPubsub()

		if config.Current.OwnerChatUser.DiscordId != "" && room.DiscordChannel != "" {
			RegisterAdmin(room.ID)
		}
	}
}

func GetAllRooms() []ChatRoom {
	defer mutex.Unlock()
	mutex.Lock()
	return lo.Map(roomKeys, func(key string, _ int) ChatRoom {
		return rooms[key]
	})
}

func GetSingleRoom(id string) (ChatRoom, bool) {
	defer mutex.Unlock()
	mutex.Lock()
	r, f := rooms[id]
	return r, f
}

func FindRoomIdByDiscordChannel(channel string) (string, bool) {
	roomIds := lo.Keys(rooms)
	return lo.Find(roomIds, func(id string) bool {
		return rooms[id].DiscordChannel == channel
	})
}

func GetUserMessageList(combinedId string) ([]*ChatMessage, bool) {
	defer mutex.Unlock()
	mutex.Lock()
	m, f := userMessages[combinedId]
	return m, f
}

func GetMessagesByUser(combinedId string) ([]*ChatMessage, bool) {
	messages, found := userMessages[combinedId]
	if !found {
		return make([]*ChatMessage, 0), false
	}
	ms := lo.Filter(messages, func(m *ChatMessage, _ int) bool { return m.From == combinedId })
	return ms, true
}

func GetAllUsers() []ChatUser {
	defer mutex.Unlock()
	mutex.Lock()
	return lo.Map(lo.Keys(users), func(key string, _ int) ChatUser { return users[key] })
}

func GetRoomUsers(roomId string) []ChatUser {
	defer mutex.Unlock()
	mutex.Lock()
	return lo.Map(roomUsers[roomId],
		func(combinedId string, _ int) ChatUser {
			return users[combinedId]
		},
	)
}

func GetRoomOnlineUsers(roomId string) []ChatUser {
	return lo.Filter(GetRoomUsers(roomId),
		func(user ChatUser, _ int) bool {
			return user.IsAdmin || !user.IsDiscordUser()
		},
	)
}

func GetUser(combinedId string) (ChatUser, bool) {
	defer mutex.Unlock()
	mutex.Lock()
	u, f := users[combinedId]
	return u, f
}

func GetUserByNickname(nickname string) (ChatUser, bool) {
	defer mutex.Unlock()
	mutex.Lock()
	cleanNickname := strings.TrimSpace(strings.ToLower(nickname))

	var user ChatUser
	for _, user = range users {
		if strings.TrimSpace(strings.ToLower(user.Nickname)) == cleanNickname {
			return user, true
		}
	}

	return user, false
}

func GetUserByDiscordId(discordId string) (ChatUser, bool) {
	defer mutex.Unlock()
	mutex.Lock()

	var user ChatUser
	for _, user = range users {
		if user.DiscordId == discordId {
			return user, true
		}
	}

	return user, false
}

func RegisterUser(user ChatUser) (string, error) {
	err := instantiateUser(user)

	if err != nil {
		return "", err
	}

	userMessages[user.ID] = append(userMessages[user.ID], roomMessageHistory[user.RoomId]...)

	if user.DiscordId == "" {
		SendMessage(&ChatMessage{
			RoomID:               user.RoomId,
			Time:                 time.Now().UTC(),
			Message:              "{nickname} has joined the room!",
			IsSystemMessage:      true,
			SystemMessageSubject: &user,
			Privately:            false,
			SpeechMode:           SPEECH_MODES[0].Value,
			From:                 user.ID,
			To:                   "",
			InvolvedUsers:        []ChatUser{user},
		})

		room, found := GetSingleRoom(user.RoomId)
		if found && room.IntroMessage != "" && user.IsWebUser {
			SendMessage(&ChatMessage{
				RoomID:               room.ID,
				Time:                 time.Now().UTC(),
				Message:              room.IntroMessage,
				IsSystemMessage:      true,
				SystemMessageSubject: &user,
				Privately:            true,
				SpeechMode:           SPEECH_MODES[0].Value,
				From:                 user.ID,
				To:                   user.ID,
				InvolvedUsers:        []ChatUser{user},
			})
		}

		userListUpdated(user.RoomId, ChatUserJoinedEvent{User: user})
	}

	return user.ID, nil
}

func DeregisterUser(combinedId string) {
	user := removeUser(combinedId)

	if user.ID != "" && user.DiscordId == "" {
		SendMessage(&ChatMessage{
			RoomID:               user.RoomId,
			Time:                 time.Now().UTC(),
			Message:              "{nickname} has left the room!",
			IsSystemMessage:      true,
			SystemMessageSubject: &user,
			Privately:            false,
			SpeechMode:           SPEECH_MODES[0].Value,
			From:                 user.ID,
			To:                   "",
			InvolvedUsers:        []ChatUser{user},
		})

		userListUpdated(user.RoomId, ChatUserLeftEvent{User: user})
	}

}

func SendMessage(message *ChatMessage) {
	defer mutex.Unlock()
	mutex.Lock()

	for _, combinedId := range roomUsers[message.RoomID] {
		if message.Privately && message.To != "" && (message.To != combinedId && message.From != combinedId) {
			continue
		}

		userMessages[combinedId] = append(userMessages[combinedId], message)
	}

	// Keeps a brief history of the public messages in the room so new people who login
	// See some activity on the chat.
	if !message.Privately || message.To == "" {
		messages := roomMessageHistory[message.RoomID]
		if len(roomMessageHistory[message.RoomID]) >= MAX_ROOM_MESSAGE_HISTORY {
			initial := len(messages) - MAX_ROOM_MESSAGE_HISTORY + 1
			messages = messages[initial:]
		}

		roomMessageHistory[message.RoomID] = append(messages, message)
	}

	RoomEvents[message.RoomID].Publish(ChatMessageEvent{Message: message})
}

func Ping(combinedId string) {
	defer mutex.Unlock()
	mutex.Lock()
	_, found := userPings[combinedId]
	if !found {
		return
	}

	userPings[combinedId] = time.Now().UTC()
}

func IsUserStale(combinedId string) bool {
	defer mutex.Unlock()
	mutex.Lock()
	_, found := userPings[combinedId]
	if !found {
		return false
	}

	lastPing := userPings[combinedId]

	return time.Now().UTC().Sub(lastPing).Seconds() > USER_STALE_TIMEOUT
}

func HasUserListChanged(combinedId string) bool {
	defer mutex.Unlock()
	mutex.Lock()
	user := users[combinedId]
	lastUserChange := userLastUserListChange[combinedId]
	lastRoomChange := roomLastUserListChange[user.RoomId]

	userLastUserListChange[combinedId] = time.Now().UTC()

	return lastRoomChange.Sub(lastUserChange).Milliseconds() > 0
}

func UserHasMessages(combinedId string) bool {
	defer mutex.Unlock()
	mutex.Lock()

	lastPing, hasPing := userPings[combinedId]

	if !hasPing {
		return false
	}

	messages, found := userMessages[combinedId]

	if !found || len(messages) == 0 {
		return false
	}

	lastMessage := messages[len(messages)-1]

	return lastMessage.Time.Sub(lastPing).Seconds() > 0
}
