package sockets

import (
	"retro-chat-rooms/chat"
)

type ISocket interface {
	GetUser() chat.ChatUser
	SetUser(user chat.ChatUser)
	GetState(key string) interface{}
	SetState(key string, value interface{})
	GetClientIP() string
	Read() ([]byte, error)
	Write(string) error
	Close() error
	ID() string
}
