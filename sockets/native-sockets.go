package sockets

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"retro-chat-rooms/chat"
	"retro-chat-rooms/pubsub"
	"sync"
	"syscall"

	"github.com/google/uuid"
)

type NativeSockets struct {
	conn   net.Conn
	id     string
	user   chat.ChatUser
	states map[string]interface{}
	mutex  sync.Mutex
}

func NewNativeSockets(id string, conn net.Conn) NativeSockets {
	return NativeSockets{
		id:     id,
		conn:   conn,
		states: make(map[string]interface{}),
		mutex:  sync.Mutex{},
		user:   chat.ChatUser{},
	}
}

func (ns *NativeSockets) Read() ([]byte, error) {
	buf := make([]byte, 1024)
	n, err := ns.conn.Read(buf)
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}

func (ns *NativeSockets) Write(msg []byte) error {
	_, err := ns.conn.Write(msg)
	return err
}

func (ns *NativeSockets) Close() error {
	return ns.conn.Close()
}

func (ns *NativeSockets) ID() string {
	return ns.id
}

func (ns *NativeSockets) GetUser() chat.ChatUser {
	return ns.user
}

func (ns *NativeSockets) SetUser(user chat.ChatUser) {
	ns.user = user
}

func (ns *NativeSockets) GetState(key string) interface{} {
	ns.mutex.Lock()
	defer ns.mutex.Unlock()

	return ns.states[key]
}

func (ns *NativeSockets) SetState(key string, value interface{}) {
	ns.mutex.Lock()
	defer ns.mutex.Unlock()

	ns.states[key] = value
}

func (ns *NativeSockets) GetClientIP() string {
	clientIP, _, err := net.SplitHostPort(ns.conn.RemoteAddr().String())
	if err != nil {
		return ""
	}

	return clientIP
}

func observeRoomEvents(connection ISocket, events pubsub.Pubsub, ctx context.Context) {
	c := events.Subscribe(connection.ID())
	defer func() {
		events.Unsubscribe(connection.ID())
	}()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Context canceled in observeRoomEvents")
			return
		case message := <-c:
			switch evt := message.(type) {

			case chat.ChatMessageEvent:
				PushMessage(connection, *evt.Message, false)

			case chat.ChatUserJoinedEvent:
				PushUserJoined(connection, evt.User)

			case chat.ChatUserLeftEvent:
				PushUserLeft(connection, evt.User)

			case chat.ChatUserKickedEvent:
				PushUserKickedMessage(connection, evt)

			}
		}
	}
}

func processClient(connection ISocket) {
	defer connection.Close()
	fmt.Printf("Processing client %s\n", connection.ID())

	userRegistered := make(chan chat.ChatUser)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		defer fmt.Println("Exiting user registration goroutine")
		c := InternalEvents.Subscribe("conn_" + connection.ID())
		defer InternalEvents.Unsubscribe("conn_" + connection.ID())
		defer close(userRegistered)
		for {
			select {
			case <-ctx.Done():
				fmt.Println("Context canceled for user registration")
				return
			case message := <-c:
				switch msg := message.(type) {
				case InternalUserRegisteredEvent:
					if connection.ID() == msg.ConnectionID {
						connection.SetUser(msg.User)
						userRegistered <- msg.User
						return
					}
				}
			}
		}
	}()

	go func() {
		defer fmt.Println("Exiting room observation goroutine")
		select {
		case <-ctx.Done():
			fmt.Println("Context canceled for room observation")
			return
		case registeredUser := <-userRegistered:
			roomEvents := chat.RoomEvents[registeredUser.RoomId]
			observeRoomEvents(connection, roomEvents, ctx)
		}
	}()

	for {
		msg, err := connection.Read()
		if err != nil {
			if err == io.EOF || err == syscall.EPIPE {
				fmt.Println("Client disconnected:", err.Error())

				user := connection.GetUser()
				chat.DeregisterUser(user.ID)
				cancel()
				break
			}
			fmt.Println("Error reading:", err.Error())
			cancel()
			break
		}

		fmt.Println("Received: ", string(msg))

		ProcessMessage(connection, msg)
	}
}

func ServeSockets(serverPort string) {
	server, err := net.Listen("tcp", "0.0.0.0:"+serverPort)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	defer server.Close()
	fmt.Println("Listening Sockets on " + "0.0.0.0:" + serverPort)
	fmt.Println("Waiting for Native client...")

	for {
		connection, err := server.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			continue
		}

		clientID := uuid.New().String()
		conn := NewNativeSockets(clientID, connection)

		fmt.Printf("Client %s connected\n", clientID)

		go processClient(&conn)
	}
}
