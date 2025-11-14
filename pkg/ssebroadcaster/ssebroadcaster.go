package ssebroadcaster

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

var (
	connPool ConnPool = NewConnPool()
)

func channelSubHandler(w http.ResponseWriter, r *http.Request, c *ConnChan, heartBeat int) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	ctx := r.Context()
	log.Printf("listener connected: id=%s conn=%d", c.Id, c.ConnId)

	ticker := time.NewTicker(time.Duration(heartBeat) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case msg, ok := <-c.Channel:
			if !ok {
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		case <-ticker.C:
			fmt.Fprintf(w, ": heartbeat\n\n")
			flusher.Flush()
		case <-ctx.Done():
			log.Printf("client disconnected: id=%s conn=%d", c.Id, c.ConnId)
			connPool.RemoveChannel(c.ConnId)

			close(c.Channel)
			return
		}
	}
}

func SseConnHandler(w http.ResponseWriter, r *http.Request, resourceId string, heartBeat int) {
	connId := uuid.New().ID()
	ch := make(chan string, 1)
	c := &ConnChan{
		ConnId:  connId,
		Id:      resourceId,
		Channel: ch,
	}
	connPool.AddChannel(c)
	channelSubHandler(w, r, c, heartBeat)
}

func BroadcastHandler(w http.ResponseWriter, r *http.Request, resourceId string) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	bodyStr := string(body)

	connPool.Broadcast(resourceId, bodyStr)
}
