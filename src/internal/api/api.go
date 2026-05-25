package api

import (
	"encoding/json"
	"net/http"
	"time"

	"media-jukebox-backend/internal/model"
	"media-jukebox-backend/internal/queue"
	"media-jukebox-backend/internal/session"
	"media-jukebox-backend/internal/ws"
)

type API struct {
	Queue   *queue.Queue
	Session *session.Manager
	Hub     *ws.Hub
}

type AddRequest struct {
	URL   string `json:"url"`
	Title string `json:"title"`
}

func generateID() string {
	return time.Now().Format("20060102150405.000000")
}

func (a *API) broadcastQueue() {
	a.Hub.Broadcast(map[string]any{
		"type":  "queue",
		"items": a.Queue.Items(),
	})
}

func (a *API) broadcastSession() {
	a.Hub.Broadcast(map[string]any{
		"type":    "session",
		"session": a.Session.Get(),
	})
}

func (a *API) Add(w http.ResponseWriter, r *http.Request) {
	var req AddRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	item := model.QueueItem{
		ID:    generateID(),
		Title: req.Title,
		URL:   req.URL,
	}

	if item.Title == "" {
		item.Title = item.URL
	}

	a.Queue.Add(item)
	a.broadcastQueue()

	w.WriteHeader(http.StatusCreated)
}

func (a *API) GetQueue(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(a.Queue.Items())
}

func (a *API) GetSession(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(a.Session.Get())
}

func (a *API) Play(w http.ResponseWriter, r *http.Request) {
	item := a.Queue.Start()

	if item == nil {
		http.Error(w, "queue empty", http.StatusBadRequest)
		return
	}

	a.Session.SetPlaying(item.ID)
	a.broadcastSession()

	a.Hub.Broadcast(map[string]any{
		"type": "play",
		"item": item,
	})
}

func (a *API) Pause(w http.ResponseWriter, r *http.Request) {
	a.Session.Pause()
	a.broadcastSession()

	a.Hub.Broadcast(map[string]any{
		"type": "pause",
	})
}

func (a *API) Next(w http.ResponseWriter, r *http.Request) {
	a.advance()
}

func (a *API) HandleEnded(itemID string) {
	current := a.Queue.Current()

	if current == nil {
		return
	}

	if current.ID != itemID {
		return
	}

	a.advance()
}

func (a *API) advance() {
	item := a.Queue.Next()

	if item == nil {
		a.Session.Idle()
		a.broadcastSession()

		a.Hub.Broadcast(map[string]any{
			"type": "idle",
		})

		return
	}

	a.Session.SetPlaying(item.ID)
	a.broadcastSession()

	a.Hub.Broadcast(map[string]any{
		"type": "play",
		"item": item,
	})
}
