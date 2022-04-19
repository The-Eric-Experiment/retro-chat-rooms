package chatroom

import (
	"errors"
	"retro-chat-rooms/config"
	"retro-chat-rooms/pubsub"
	"retro-chat-rooms/session"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/samber/lo"
)

type Room struct {
	ID                 string
	Name               string
	Description        string
	Color              string
	DiscordChannel     string
	Messages           []*RoomMessage
	Users              []*RoomUser
	DiscordUsers       []*RoomUser
	LastUserListUpdate time.Time
	mutex              sync.Mutex
	ChatEventAwaiter   ChatEventAwaiter
	TextColor          string
}

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

func (room *Room) Initialize() {
	room.TextColor = determineTextColor(room.Color)
	room.ChatEventAwaiter.Initialize()
}

func (room *Room) RegisterUser(nickname string, color string, sessionIdent string) (*RoomUser, error) {
	now := time.Now().UTC()
	inputNickname := strings.ToLower(strings.TrimSpace(nickname))

	userId := uuid.NewSHA1(uuid.NameSpaceURL, []byte(inputNickname)).String()

	room.mutex.Lock()
	_, found := lo.Find(room.Users, func(user *RoomUser) bool { return user.ID == userId })
	room.mutex.Unlock()

	if found {
		return nil, errors.New("user exists")
	}

	user := &RoomUser{
		ID:                 userId,
		LastActivity:       now,
		Nickname:           nickname,
		Color:              color,
		LastPing:           now,
		LastUserListUpdate: now,
		SessionIdent:       sessionIdent,
	}

	room.mutex.Lock()
	room.Users = append(room.Users, user)
	room.mutex.Unlock()

	room.SendMessage(&RoomMessage{
		Time:            time.Now().UTC(),
		Message:         "{nickname} has joined the room!",
		IsSystemMessage: true,
		Privately:       false,
		SpeechMode:      SPEECH_MODES[0].Value,
		From:            user,
	})

	room.LastUserListUpdate = time.Now().UTC()

	room.Update()

	return user, nil
}

func (room *Room) RegisterDiscordUser(discordUser *discordgo.User) *RoomUser {
	now := time.Now().UTC()

	room.mutex.Lock()
	foundUser, found := lo.Find(room.DiscordUsers, func(user *RoomUser) bool { return user.ID == discordUser.ID })
	room.mutex.Unlock()

	if found {
		return foundUser
	}

	user := &RoomUser{
		ID:                 discordUser.ID,
		LastActivity:       now,
		Nickname:           discordUser.Username,
		Color:              USER_COLOR_BLACK,
		LastPing:           now,
		LastUserListUpdate: now,
		SessionIdent:       discordUser.ID,
		IsOwner:            false,
		IsDiscordUser:      true,
	}

	room.mutex.Lock()
	room.DiscordUsers = append(room.DiscordUsers, user)
	room.mutex.Unlock()

	room.Update()

	return user
}

func (room *Room) DeregisterUser(user *RoomUser) {
	if user.ID == OwnerRoomUser.ID {
		LogoutOwner()
		return
	}

	room.mutex.Lock()
	room.Users = lo.Filter(room.Users, func(u *RoomUser, _ int) bool {
		return u.ID != user.ID
	})
	room.mutex.Unlock()

	room.SendMessage(&RoomMessage{
		Time:            time.Now().UTC(),
		Message:         "{nickname} has left the room!",
		IsSystemMessage: true,
		Privately:       false,
		SpeechMode:      SPEECH_MODES[0].Value,
		From:            user,
	})

	room.LastUserListUpdate = time.Now().UTC()

	room.Update()
}

func (room *Room) DeregisterDiscordUser(user *RoomUser) {
	room.mutex.Lock()
	room.Users = lo.Filter(room.DiscordUsers, func(u *RoomUser, _ int) bool {
		return u.ID != user.ID
	})
	room.mutex.Unlock()

	room.Update()
}

func (room *Room) SendMessage(message *RoomMessage) {
	defer room.mutex.Unlock()
	room.mutex.Lock()
	room.Messages = append([]*RoomMessage{message}, room.Messages...)

	pubsub.Messaging.Publish(&PubSubPostMessage{
		Room:    room,
		Message: message,
	})

	if checkForMessageAbuse(room, message.From) {
		session.SetFlooded(message.From.SessionIdent)
	}

	if len(room.Messages) > MAX_MESSAGES {
		room.Messages = room.Messages[0:MAX_MESSAGES]
	}
}

func (room *Room) GetUser(userId string) *RoomUser {
	defer room.mutex.Unlock()
	room.mutex.Lock()
	if OwnerRoomUser != nil && (userId == OwnerRoomUser.ID || userId == config.Current.OwnerChatUser.DiscordId) {
		return OwnerRoomUser
	}

	discordUser, found := lo.Find(room.DiscordUsers, func(r *RoomUser) bool { return r.ID == userId })

	if found {
		return discordUser
	}

	roomUser, found := lo.Find(room.Users, func(r *RoomUser) bool { return r.ID == userId })
	if found {
		return roomUser
	}

	return nil
}

func (room *Room) GetUserByNickname(nickname string) *RoomUser {
	cleanNickname := strings.TrimSpace(strings.ToLower(nickname))
	defer room.mutex.Unlock()
	room.mutex.Lock()

	if cleanNickname == strings.TrimSpace(strings.ToLower(OwnerRoomUser.Nickname)) {
		return OwnerRoomUser
	}

	discordUser, found := lo.Find(room.DiscordUsers, func(r *RoomUser) bool { return strings.TrimSpace(strings.ToLower(r.Nickname)) == cleanNickname })

	if found {
		return discordUser
	}

	roomUser, found := lo.Find(room.Users, func(r *RoomUser) bool { return strings.TrimSpace(strings.ToLower(r.Nickname)) == cleanNickname })
	if found {
		return roomUser
	}

	return nil
}

func (room *Room) GetUserBySessionIdent(sessionIdent string) *RoomUser {
	defer room.mutex.Unlock()
	room.mutex.Lock()

	if sessionIdent == OwnerRoomUser.SessionIdent {
		return OwnerRoomUser
	}

	discordUser, found := lo.Find(room.DiscordUsers, func(r *RoomUser) bool { return r.SessionIdent == sessionIdent })

	if found {
		return discordUser
	}

	roomUser, found := lo.Find(room.Users, func(r *RoomUser) bool { return r.SessionIdent == sessionIdent })
	if found {
		return roomUser
	}

	return nil
}

func (room *Room) Update() {
	room.ChatEventAwaiter.Dispatch()
}

func FromConfig() []*Room {
	return lo.Map(config.Current.Rooms, func(cr config.ConfigChatRoom, _ int) *Room {
		room := &Room{
			mutex:              sync.Mutex{},
			ChatEventAwaiter:   ChatEventAwaiter{},
			ID:                 cr.ID,
			Name:               cr.Name,
			Description:        cr.Description,
			Color:              cr.Color,
			DiscordChannel:     cr.DiscordChannel,
			Messages:           make([]*RoomMessage, 0),
			Users:              make([]*RoomUser, 0),
			DiscordUsers:       make([]*RoomUser, 0),
			LastUserListUpdate: time.Now().UTC(),
			TextColor:          determineTextColor(cr.Color),
		}

		room.Initialize()
		return room
	})
}

func checkForMessageAbuse(room *Room, user *RoomUser) bool {
	if user.IsDiscordUser || user.IsOwner {
		return false
	}

	messages := lo.Filter(room.Messages, func(message *RoomMessage, _ int) bool { return message.From.ID == user.ID })

	if len(messages) < session.MESSAGE_FLOOD_MESSAGES_COUNT {
		return false
	}

	firstMessage, err := lo.Last(messages[0:session.MESSAGE_FLOOD_MESSAGES_COUNT])
	if err != nil {
		return false
	}

	messages = messages[0 : session.MESSAGE_FLOOD_MESSAGES_COUNT-1]

	for _, msg := range messages {
		seconds := msg.Time.Sub(firstMessage.Time).Seconds()
		if seconds > session.MESSAGE_FLOOD_RANGE_SEC {
			return false
		}
	}

	return true
}
