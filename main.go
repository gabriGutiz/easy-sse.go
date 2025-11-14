package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/gabriGutiz/easy-sse.go/pkg/ssebroadcaster"
)

func main() {
	addrFlag := flag.String("addr", ":8080", "The address to listen on")
	heartBeatFlag := flag.Int("heart-beat-ms", 10000, "The heart beat time in ms to send to subscribers")
	flag.Parse()
	addr := *addrFlag
	heartBeat := *heartBeatFlag

	http.HandleFunc("/channels/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		ssebroadcaster.SseConnHandler(w, r, heartBeat)
	})

	http.HandleFunc("/channels/{id}/broadcast", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		ssebroadcaster.BroadcastHandler(w, r)
	})

	log.Printf("Listening on %s\n", addr)
	log.Printf("Subscribers heart beat %dms\n", heartBeat)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatalf("Error initializing server: %s", err)
	}
}
