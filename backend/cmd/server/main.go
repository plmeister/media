package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"media/internal/api"
	"media/internal/identity"
	"media/internal/mpv"
	"media/internal/obs"
	"media/internal/queue"
	"media/internal/session"
	"media/internal/ui"
	"media/internal/ws"

	"go.opentelemetry.io/otel"
)

func setupFrontend(mode string) http.HandlerFunc {
	switch mode {
	case "dev":
		return func(w http.ResponseWriter, r *http.Request) {
			proxy := httputil.NewSingleHostReverseProxy(&url.URL{
				Scheme: "http",
				Host:   "localhost:5173",
			})
			proxy.ServeHTTP(w, r)
		}
	}

	return ui.Handler()
}

func otelMiddleware(next http.HandlerFunc) http.HandlerFunc {
	tracer := otel.Tracer("http")

	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), r.Method+" "+r.URL.Path)
		defer span.End()

		// propagate request into handler
		next(w, r.WithContext(ctx))
	}
}

func main() {
	q := queue.New()
	s := session.New()
	hub := ws.New()
	logger := obs.NewLogger("backend")

	ctx := context.Background()
	tp, err := obs.InitNoOpTracer(ctx, "backend")
	if err != nil {
		panic(err)
	}
	defer func() { _ = tp.Shutdown(ctx) }()
	logger.Info(ctx, "backend starting", nil)
	frontEndMode := os.Getenv("FRONTEND_MODE")

	id, err := identity.LoadOrCreate("./device.json", "player")
	if err != nil {
		panic(err)
	}

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

	player.SetStateChangeHandler(func() {
		apiHandler.BroadcastState()
	})

	hub.OnMessage = func(msg ws.ClientMessage) {
		switch msg.Type {
		case "ended":
			apiHandler.HandleEnded(msg.ItemID)
		}
	}

	go func() {
		for {
			e := <-player.Events()
			logger.Info(ctx, "player_event", map[string]any{
				"name": e.Name,
				"data": e.Data,
			})

		}
	}()

	http.HandleFunc("/ws", logging(hub.HandleWS))

	apiMux := http.NewServeMux()

	apiMux.HandleFunc("/state", otelMiddleware(logging(apiHandler.State)))
	apiMux.HandleFunc("/queue/add", otelMiddleware(logging(apiHandler.Add)))
	apiMux.HandleFunc("/queue/clear", otelMiddleware(logging(apiHandler.Clear)))
	apiMux.HandleFunc("/queue", otelMiddleware(logging(apiHandler.GetQueue)))

	apiMux.HandleFunc("/control/play", otelMiddleware(logging(apiHandler.Play)))
	apiMux.HandleFunc("/control/pause", otelMiddleware(logging(apiHandler.Pause)))
	apiMux.HandleFunc("/control/resume", otelMiddleware(logging(apiHandler.Resume)))
	apiMux.HandleFunc("/control/next", otelMiddleware(logging(apiHandler.Next)))
	apiMux.HandleFunc("/control/prev", otelMiddleware(logging(apiHandler.Prev)))
	apiMux.HandleFunc("/identity", otelMiddleware(logging(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(map[string]any{
			"id":    id.ID,
			"label": id.Label,
			"role":  id.Role,
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})))
	apiMux.HandleFunc("/session", otelMiddleware(logging(apiHandler.GetSession)))

	http.Handle("/api/", http.StripPrefix("/api", apiMux))
	http.HandleFunc("/", otelMiddleware(logging(setupFrontend(frontEndMode))))

	logger.Info(ctx, "backend listening on :8080", map[string]any{})
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		logger.Error(ctx, "Error", err, map[string]any{})
	}
}

func logging(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			reqID = generateID()
		}

		ctx = obs.WithRequestID(ctx, reqID)

		logger := obs.NewLogger("http")

		logger.Info(ctx, "request", map[string]any{
			"method":     r.Method,
			"path":       r.URL.Path,
			"remote":     r.RemoteAddr,
			"request_id": reqID,
		})

		w.Header().Set("X-Request-ID", reqID)
		next(w, r.WithContext(ctx))
	}
}

func generateID() string {
	// simple, dependency-free request id
	b := make([]byte, 16)
	_, _ = rand.Read(b)

	return hex.EncodeToString(b)
}
