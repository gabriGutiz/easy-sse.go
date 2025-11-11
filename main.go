package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

type ConnChan struct {
	ConnId  uint32
	Id      string
	Channel chan string
}

var (
	chans   []*ConnChan
	chansMu sync.Mutex
)

func addChannel(c *ConnChan) {
	chansMu.Lock()
	chans = append(chans, c)
	chansMu.Unlock()
}

func removeChannel(connId uint32) {
	chansMu.Lock()
	defer chansMu.Unlock()

	for i, c := range chans {
		if c.ConnId == connId {
			chans[i] = chans[len(chans)-1]
			chans = chans[:len(chans)-1]
			return
		}
	}
}

func main() {
	addrFlag := flag.String("addr", ":8080", "The address to listen on")
	heartBeatFlag := flag.Int("heart-beat-ms", 10000, "The heart beat time in ms to send to subscribers")
	flag.Parse()
	addr := *addrFlag
	heartBeat := *heartBeatFlag

	http.HandleFunc("/channels/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		connId := uuid.New().ID()
		ch := make(chan string, 1)
		c := &ConnChan{
			ConnId:  connId,
			Id:      id,
			Channel: ch,
		}
		addChannel(c)
		channelSubHandler(w, r, c, heartBeat)
	})

	http.HandleFunc("/channels/{id}/broadcast", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()
		bodyStr := string(body)

		chansMu.Lock()
		for _, c := range chans {
			if c.Id == id {
				select {
				case c.Channel <- bodyStr:
					log.Printf("sent to %s: %s", id, bodyStr)
				default:
					log.Printf("channel busy for %s, dropping message", id)
				}
			}
		}
		chansMu.Unlock()
	})

	log.Printf("Listening on %s\n", addr)
	log.Printf("Subscribers heart beat %dms\n", heartBeat)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatalf("Error initializing server: %s", err)
	}
}

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
			removeChannel(c.ConnId)

			close(c.Channel)
			return
		}
	}
}
