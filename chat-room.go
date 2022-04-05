package main

import (
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/samber/lo"
)

const MAX_MESSAGES = 200
const USER_LOGOUT_TIMEOUT = 120
const USER_MAX_MESSAGE_RATE_SEC = 60
const USER_MIN_MESSAGE_RATE_SEC = 2
const USER_SCREAM_TIMEOUT_MIN float64 = 2
const CHILL_OUT_TIMEOUT_MIN float64 = 2
const UPDATER_WAIT_TIMEOUT_MS = 30000

type SpeechMode struct {
	Value string
	Label string
}

type RoomMessage struct {
	Time            time.Time
	Message         string
	From            *RoomUser
	To              *RoomUser
	Privately       bool
	SpeechMode      string
	IsSystemMessage bool
}

type RoomUser struct {
	ID                 string
	LastActivity       time.Time
	Nickname           string
	Color              string
	LastPing           time.Time
	LastUserListUpdate time.Time
	SessionIdent       string
}

type Room struct {
	ID                 string `yaml:"id"`
	Name               string `yaml:"name"`
	Description        string `yaml:"description"`
	Color              string `yaml:"color"`
	Messages           []*RoomMessage
	Users              []*RoomUser
	LastUserListUpdate time.Time
	mutex              sync.Mutex
	chatEventAwaiter   ChatEventAwaiter
}

func (room *Room) RegisterUser(nickname string, color string, sessionIdent string) (*RoomUser, error) {
	now := time.Now().UTC()
	inputNickname := strings.ToLower(strings.TrimSpace(nickname))

	room.mutex.Lock()
	_, found := lo.Find(room.Users, func(user *RoomUser) bool { return user.Nickname == inputNickname })
	room.mutex.Unlock()

	if found {
		return nil, errors.New("user exists")
	}

	userId := uuid.NewSHA1(uuid.NameSpaceURL, []byte(inputNickname)).String()

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
		SpeechMode:      speechModes[0].Value,
		From:            user,
	})

	room.LastUserListUpdate = time.Now().UTC()

	room.Update()

	return user, nil
}

func (room *Room) DeregisterUser(user *RoomUser) {
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
		SpeechMode:      speechModes[0].Value,
		From:            user,
	})

	room.LastUserListUpdate = time.Now().UTC()

	room.Update()
}

func (room *Room) SendMessage(message *RoomMessage) {
	defer room.mutex.Unlock()
	room.mutex.Lock()
	room.Messages = append([]*RoomMessage{message}, room.Messages...)
	if len(room.Messages) > MAX_MESSAGES {
		room.Messages = room.Messages[0:MAX_MESSAGES]
	}
}

func (room *Room) GetUser(userId string) (*RoomUser, bool) {
	defer room.mutex.Unlock()
	room.mutex.Lock()
	return lo.Find(room.Users, func(r *RoomUser) bool { return r.ID == userId })
}

func (room *Room) Update() {
	room.chatEventAwaiter.Dispatch()
}
