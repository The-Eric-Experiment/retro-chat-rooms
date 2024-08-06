package sockets

import "retro-chat-rooms/chat"

type ISocket interface {
	GetUser() chat.ChatUser
	SetUser(user chat.ChatUser)
	Read() ([]byte, error)
	Write(string) error
	Close() error
	ID() string
}
