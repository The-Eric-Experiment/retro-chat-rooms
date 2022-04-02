package main

import (
	"errors"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/samber/lo"
)

const MAX_MESSAGES = 200
const USER_LOGOUT_TIMEOUT = 120
const USER_MAX_MESSAGE_RATE_SEC = 60
const USER_MIN_MESSAGE_RATE_SEC = 2

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
	UpdateMessageRate  int
}

type Room struct {
	ID                 string `yaml:"id"`
	Name               string `yaml:"name"`
	Color              string `yaml:"color"`
	Messages           []*RoomMessage
	Users              []*RoomUser
	LastUserListUpdate time.Time
}

func getMessageRate(room *Room, user *RoomUser) int {
	if len(room.Messages) == 0 {
		return USER_MAX_MESSAGE_RATE_SEC
	}

	first := room.Messages[0]

	now := time.Now().UTC()

	//UPDATE FAST FOR TWO MINUTES
	diff := math.Floor(now.Sub(first.Time).Seconds() / 2)

	if diff > USER_MAX_MESSAGE_RATE_SEC {
		return USER_MAX_MESSAGE_RATE_SEC
	}

	if diff < USER_MIN_MESSAGE_RATE_SEC {
		return USER_MIN_MESSAGE_RATE_SEC
	}

	return int(diff)
}

func UpdateMessageRate(room *Room, user *RoomUser) {
	user.UpdateMessageRate = getMessageRate(room, user)
}

func (room *Room) RegisterUser(nickname string, color string) (*RoomUser, error) {
	inputNickname := strings.ToLower(strings.TrimSpace(nickname))
	_, found := lo.Find(room.Users, func(user *RoomUser) bool { return user.Nickname == inputNickname })

	if found {
		return nil, errors.New("user exists")
	}

	userId := uuid.NewSHA1(uuid.NameSpaceURL, []byte(inputNickname)).String()

	user := &RoomUser{
		ID:                 userId,
		LastActivity:       time.Now().UTC(),
		Nickname:           nickname,
		Color:              color,
		LastPing:           time.Now().UTC(),
		LastUserListUpdate: time.Now().UTC(),
	}

	room.Users = append(room.Users, user)

	room.SendMessage(&RoomMessage{
		Time:            time.Now().UTC(),
		Message:         "{nickname} has joined the room!",
		IsSystemMessage: true,
		Privately:       false,
		SpeechMode:      speechModes[0].Value,
		From:            user,
	})

	room.LastUserListUpdate = time.Now().UTC()

	return user, nil
}

func (room *Room) DeregisterUser(user *RoomUser) {
	room.Users = lo.Filter(room.Users, func(u *RoomUser, _ int) bool {
		return u.ID != user.ID
	})

	room.SendMessage(&RoomMessage{
		Time:            time.Now().UTC(),
		Message:         "{nickname} has left the room!",
		IsSystemMessage: true,
		Privately:       false,
		SpeechMode:      speechModes[0].Value,
		From:            user,
	})

	room.LastUserListUpdate = time.Now().UTC()
}

func (room *Room) SendMessage(message *RoomMessage) {
	room.Messages = append([]*RoomMessage{message}, room.Messages...)
	if len(room.Messages) > MAX_MESSAGES {
		room.Messages = room.Messages[0:MAX_MESSAGES]
	}
}
