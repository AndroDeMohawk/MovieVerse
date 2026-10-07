package ws

import (
	"encoding/json"
	"log/slog"
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	Hub     *Hub
	Conn    *websocket.Conn
	Send    chan []byte
	MovieID int64
}

type Hub struct {
	// Комнаты по movie_id: map[movieID]map[*Client]bool
	rooms      map[int64]map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan BroadcastMessage
	mu         sync.RWMutex
	log        *slog.Logger
}

type BroadcastMessage struct {
	MovieID int64
	Payload []byte
}

func NewHub(log *slog.Logger) *Hub {
	return &Hub{
		rooms:      make(map[int64]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan BroadcastMessage, 256),
		log:        log,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if _, exists := h.rooms[client.MovieID]; !exists {
				h.rooms[client.MovieID] = make(map[*Client]bool)
			}
			h.rooms[client.MovieID][client] = true
			h.mu.Unlock()
			h.log.Debug("ws client registered", slog.Int64("movie_id", client.MovieID))

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.rooms[client.MovieID]; ok {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.Send)
					if len(clients) == 0 {
						delete(h.rooms, client.MovieID)
					}
				}
			}
			h.mu.Unlock()
			h.log.Debug("ws client unregistered", slog.Int64("movie_id", client.MovieID))

		case msg := <-h.broadcast:
			h.mu.RLock()
			clients := h.rooms[msg.MovieID]
			for client := range clients {
				select {
				case client.Send <- msg.Payload:
				default:
					close(client.Send)
					delete(clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) BroadcastToMovie(movieID int64, payload interface{}) {
	bytes, err := json.Marshal(payload)
	if err != nil {
		h.log.Error("failed to marshal ws broadcast payload", slog.String("error", err.Error()))
		return
	}

	h.broadcast <- BroadcastMessage{
		MovieID: movieID,
		Payload: bytes,
	}
}
