/* Package api */
package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"media/internal/model"
	"media/internal/mpv"
	"media/internal/queue"
	"media/internal/session"
	"media/internal/ws"
)

type API struct {
	Queue   *queue.Queue
	Session *session.Manager
	Hub     *ws.Hub
	Player  *mpv.Player
}

type AddRequest struct {
	Type  string `json:"type"`
	URL   string `json:"url"`
	Title string `json:"title"`
}

type PlayerState struct {
	Playing bool              `json:"playing"`
	Current *model.QueueItem  `json:"current"`
	Queue   []model.QueueItem `json:"queue"`
}

func generateID() string {
	return time.Now().Format("20060102150405.000000")
}

func (a *API) GetState() PlayerState {
	return PlayerState{
		Playing: a.Player.Playing,
		Current: a.Queue.Current(),
		Queue:   a.Queue.Items(),
	}
}

func (a *API) BroadcastState() {
	a.Hub.Broadcast(map[string]any{
		"type": "state",
		"data": a.GetState(),
	})
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

func (a *API) State(w http.ResponseWriter, r *http.Request) {
	if err := json.NewEncoder(w).Encode(a.GetState()); err != nil {
		_, _ = w.Write([]byte("error"))
	}
}

func (a *API) Add(w http.ResponseWriter, r *http.Request) {
	var req AddRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	log.Printf("add %s:%s to queue", req.Type, req.URL)
	item := model.QueueItem{
		ID:     generateID(),
		Type:   req.Type,
		Title:  req.Title,
		Source: req.URL,
	}

	if item.Title == "" {
		item.Title = item.Source
	}

	a.Queue.Add(item)
	a.broadcastQueue()
	a.BroadcastState()

	w.WriteHeader(http.StatusCreated)
}

func (a *API) GetQueue(w http.ResponseWriter, r *http.Request) {
	if err := json.NewEncoder(w).Encode(a.Queue.Items()); err != nil {
		_, _ = w.Write([]byte("error"))
	}
}

func (a *API) GetSession(w http.ResponseWriter, r *http.Request) {
	if err := json.NewEncoder(w).Encode(a.Session.Get()); err != nil {
		_, _ = w.Write([]byte("error"))
	}
}

func (a *API) Play(w http.ResponseWriter, r *http.Request) {
	item := a.Queue.Start()

	if item == nil {
		log.Printf("no item to play")
		http.Error(w, "queue empty", http.StatusBadRequest)
		return
	}

	a.Session.SetPlaying(item.ID)
	a.broadcastSession()

	a.Hub.Broadcast(map[string]any{
		"type": "play",
		"item": item,
	})

	_ = a.Player.LoadFile(item.Source)
	_ = a.Player.Pause(false)

	a.BroadcastState()
}

func (a *API) Pause(w http.ResponseWriter, r *http.Request) {
	a.Session.Pause()
	a.broadcastSession()

	_ = a.Player.Pause(true)
	log.Printf("set pause to true")
	a.BroadcastState()
}

func (a *API) Resume(w http.ResponseWriter, r *http.Request) {
	a.Session.Pause()
	a.broadcastSession()

	_ = a.Player.Pause(false)
	log.Printf("set pause to false")
	a.BroadcastState()
}

func (a *API) Next(w http.ResponseWriter, r *http.Request) {
	a.advance(a.Queue.Next())
}

func (a *API) Prev(w http.ResponseWriter, r *http.Request) {
	a.advance(a.Queue.Prev())
}

func (a *API) HandleEnded(itemID string) {
	current := a.Queue.Current()

	if current == nil {
		return
	}

	if current.ID != itemID {
		return
	}

	a.advance(a.Queue.Next())
	a.BroadcastState()
}

func (a *API) advance(item *model.QueueItem) {
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

	a.BroadcastState()
}

func (a *API) Clear(w http.ResponseWriter, r *http.Request) {
	_ = a.Queue.Clear()
	a.broadcastQueue()
	a.BroadcastState()
}
