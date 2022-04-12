package chatroom

import (
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ChatEventAwaiter struct {
	currentlyWaiting map[string]chan string
	mu               sync.Mutex
}

func (cea *ChatEventAwaiter) Initialize() {
	if cea.currentlyWaiting == nil {
		cea.currentlyWaiting = make(map[string]chan string, 0)
	}
}

func (cea *ChatEventAwaiter) Dispatch() {
	cea.mu.Lock()
	defer cea.mu.Unlock()
	for key, ch := range cea.currentlyWaiting {
		delete(cea.currentlyWaiting, key)
		ch <- "send"
	}
}

func (cea *ChatEventAwaiter) Await(ctx *gin.Context, cb func(ctx *gin.Context)) {
	reqId := uuid.New().String()
	fmt.Println("request received")
	cea.mu.Lock()

	ch := make(chan string)
	cea.currentlyWaiting[reqId] = ch
	fmt.Println(len(cea.currentlyWaiting))

	cea.mu.Unlock()

	go func(context *gin.Context) {
		defer close(ch)
		var wat string
		select {
		case <-ch:
			wat = "received"
			break
		case <-time.After(UPDATER_WAIT_TIMEOUT_MS * time.Millisecond):
			wat = "timedout"
			cea.mu.Lock()
			delete(cea.currentlyWaiting, reqId)
			cea.mu.Unlock()
			break
		}

		fmt.Println(wat)

		cb(context)
	}(ctx.Copy())
}
