package chatroom

import (
	"errors"
	"retro-chat-rooms/config"
	"retro-chat-rooms/pubsub"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
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

	if config.Current.OwnerChatUser.DiscordId != "" {
		room.RegisterOwner(true)
	}
}

func (room *Room) RegisterUser(userId string, nickname string, color string) (*RoomUser, error) {
	now := time.Now().UTC()

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
	foundUser, found := lo.Find(room.Users, func(user *RoomUser) bool { return user.ID == discordUser.ID && user.IsDiscordUser })
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
		IsAdmin:            false,
		IsDiscordUser:      true,
	}

	room.mutex.Lock()
	room.Users = append(room.Users, user)
	room.mutex.Unlock()

	room.Update()

	return user
}

func (room *Room) DeregisterUser(user *RoomUser) {
	if user.IsAdmin && user.IsDiscordUser {
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

func (room *Room) RegisterOwner(noMessage bool) *RoomUser {
	now := time.Now().UTC()

	cfg := config.Current.OwnerChatUser

	id := cfg.Id

	if cfg.DiscordId != "" {
		id = cfg.DiscordId
	}

	room.mutex.Lock()
	user, found := lo.Find(room.Users, func(user *RoomUser) bool { return user.ID == id })
	room.mutex.Unlock()

	if found {
		return user
	}

	user = &RoomUser{
		ID:                 id,
		LastActivity:       now,
		Nickname:           cfg.Nickname,
		Color:              cfg.Color,
		LastPing:           now,
		LastUserListUpdate: now,
		IsDiscordUser:      cfg.DiscordId != "",
		IsAdmin:            true,
	}

	room.mutex.Lock()
	room.Users = append(room.Users, user)
	room.mutex.Unlock()

	if !noMessage {
		room.SendMessage(&RoomMessage{
			Time:            time.Now().UTC(),
			Message:         "{nickname} has joined the room!",
			IsSystemMessage: true,
			Privately:       false,
			SpeechMode:      SPEECH_MODES[0].Value,
			From:            user,
		})
	}

	room.LastUserListUpdate = time.Now().UTC()

	room.Update()

	return user
}

func (room *Room) DeregisterDiscordUser(user *RoomUser) {
	room.mutex.Lock()
	room.Users = lo.Filter(room.Users, func(u *RoomUser, _ int) bool {
		return u.ID != user.ID && u.IsDiscordUser
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

	if len(room.Messages) > MAX_MESSAGES {
		room.Messages = room.Messages[0:MAX_MESSAGES]
	}
}

func (room *Room) GetUser(c *gin.Context) *RoomUser {
	session := sessions.Default(c)
	userId := session.Get("userId")

	if userId == nil {
		return nil
	}

	defer room.mutex.Unlock()
	room.mutex.Lock()

	roomUser, found := lo.Find(room.Users, func(r *RoomUser) bool { return r.ID == userId })
	if found {
		return roomUser
	}

	return nil
}

func (room *Room) GetUserByID(userId string) *RoomUser {
	defer room.mutex.Unlock()
	room.mutex.Lock()

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

	roomUser, found := lo.Find(room.Users, func(r *RoomUser) bool { return strings.TrimSpace(strings.ToLower(r.Nickname)) == cleanNickname })
	if found {
		return roomUser
	}

	return nil
}

func (room *Room) GetUserMessages(user *RoomUser) []*RoomMessage {
	defer room.mutex.Unlock()
	room.mutex.Lock()
	return lo.Filter(room.Messages, func(message *RoomMessage, _ int) bool { return message.From.ID == user.ID })
}

func (room *Room) GetOnlineUsers() int {
	defer room.mutex.Unlock()
	room.mutex.Lock()
	filtered := lo.Filter(room.Users, func(u *RoomUser, _ int) bool { return !u.IsDiscordUser })
	return len(filtered)
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
			LastUserListUpdate: time.Now().UTC(),
			TextColor:          determineTextColor(cr.Color),
		}

		room.Initialize()
		return room
	})
}
