package sockets

import (
	gorilla "github.com/gorilla/websocket"
)

type WebSocket struct {
	ws *gorilla.Conn
}

func (websocks *WebSocket) Read() ([]byte, error) {
	_, p, err := websocks.ws.ReadMessage()

	if err != nil {
		return nil, err
	}

	return p, nil
}

func (websocks *WebSocket) Write(msg string) error {
	return websocks.ws.WriteMessage(1, []byte(msg))
}

func (websocks *WebSocket) Close() error {
	return websocks.ws.Close()
}

// var upgrader = gorilla.Upgrader{
// 	ReadBufferSize:  1024,
// 	WriteBufferSize: 1024,
// 	CheckOrigin: func(r *http.Request) bool {
// 		return true
// 	},
// }

// func ServeWebSockets(c *gin.Context) {
// 	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
// 	if err != nil {
// 		fmt.Println(err.Error())
// 		return
// 	}

// 	go processClient(&WebSocket{ws: ws},)
// }
