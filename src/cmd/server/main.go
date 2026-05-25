package main

import (
	"log"
	"media-jukebox-backend/internal/api"
	"media-jukebox-backend/internal/mpv"
	"media-jukebox-backend/internal/queue"
	"media-jukebox-backend/internal/session"
	"media-jukebox-backend/internal/ws"
	"net/http"
)

func main() {
	q := queue.New()
	s := session.New()
	hub := ws.New()

	player, err := mpv.New("/tmp/mpv.sock")
	if err != nil {
		panic(err)
	}

	apiHandler := &api.API{
		Queue:   q,
		Session: s,
		Hub:     hub,
		Player:  player,
	}

	hub.OnMessage = func(msg ws.ClientMessage) {
		switch msg.Type {
		case "ended":
			apiHandler.HandleEnded(msg.ItemID)
		}
	}

	http.HandleFunc("/ws", logging(hub.HandleWS))

	http.HandleFunc("/queue/add", logging(apiHandler.Add))
	http.HandleFunc("/queue/next", logging(apiHandler.Next))
	http.HandleFunc("/queue", logging(apiHandler.GetQueue))

	http.HandleFunc("/control/play", logging(apiHandler.Play))
	http.HandleFunc("/control/pause", logging(apiHandler.Pause))
	http.HandleFunc("/control/resume", logging(apiHandler.Resume))
	http.HandleFunc("/control/next", logging(apiHandler.Next))
	http.HandleFunc("/control/prev", logging(apiHandler.Prev))

	http.HandleFunc("/session", logging(apiHandler.GetSession))

	log.Println("backend listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func logging(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf(
			"%s %s from %s",
			r.Method,
			r.URL.Path,
			r.RemoteAddr,
		)

		next(w, r)
	}
}
