package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type ClientMessage struct {
	Type   string `json:"type"`
	ItemID string `json:"itemId,omitempty"`
}

type Hub struct {
	mu        sync.Mutex
	clients   map[*websocket.Conn]bool
	upgrader  websocket.Upgrader
	OnMessage func(ClientMessage)
}

func New() *Hub {
	return &Hub{
		clients: make(map[*websocket.Conn]bool),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

func (h *Hub) HandleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade failed:", err)
		return
	}

	h.mu.Lock()
	h.clients[conn] = true
	h.mu.Unlock()

	go h.readLoop(conn)
}

func (h *Hub) readLoop(conn *websocket.Conn) {
	defer func() {
		h.mu.Lock()
		delete(h.clients, conn)
		h.mu.Unlock()
		conn.Close()
	}()

	for {
		var msg ClientMessage

		err := conn.ReadJSON(&msg)
		if err != nil {
			return
		}

		if h.OnMessage != nil {
			h.OnMessage(msg)
		}
	}
}

func (h *Hub) Broadcast(v any) {
	payload, err := json.Marshal(v)
	if err != nil {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	for conn := range h.clients {
		err := conn.WriteMessage(websocket.TextMessage, payload)
		if err != nil {
			conn.Close()
			delete(h.clients, conn)
		}
	}
}
