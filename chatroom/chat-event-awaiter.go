package chatroom

import (
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
	cea.mu.Lock()

	ch := make(chan string)
	cea.currentlyWaiting[reqId] = ch

	cea.mu.Unlock()

	go func(context *gin.Context) {
		defer close(ch)
		select {
		case <-ch:
			break
		case <-time.After(UPDATER_WAIT_TIMEOUT_MS * time.Millisecond):
			cea.mu.Lock()
			delete(cea.currentlyWaiting, reqId)
			cea.mu.Unlock()
			break
		}

		cb(context)
	}(ctx.Copy())
}
