/* Package mpv contains code for communicating with
* a running instance of mpv over a unix socket
 */
package mpv

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/otel"
)

type Command struct {
	Method string
	Args   []any
}

type Event struct {
	Name string
	Data any
}

type Response struct {
	RequestID uint64
	Error     string
	Data      any
}

type request struct {
	id      uint64
	payload []byte
}

type Player struct {
	conn    net.Conn
	mu      sync.RWMutex
	stateMu sync.RWMutex

	writeCh chan request
	events  chan Event
	pending sync.Map

	reqID    atomic.Uint64
	closeCh  chan struct{}
	closed   atomic.Bool
	playing  bool
	position float64
	posMu    sync.RWMutex

	socketPath string

	onStateChange func()
}

func (p *Player) SetStateChangeHandler(fn func()) {
	p.onStateChange = fn
}

func New(socketpath string) (*Player, error) {
	p := &Player{
		writeCh:    make(chan request, 128),
		events:     make(chan Event, 128),
		closeCh:    make(chan struct{}),
		socketPath: socketpath,
		playing:    false,
	}

	go p.reconnectLoop()
	go p.writerLoop()

	return p, nil
}

func (p *Player) reconnectLoop() {
	backoff := time.Second

	for {
		if p.closed.Load() {
			return
		}

		conn, err := net.Dial("unix", p.socketPath)
		if err != nil {
			time.Sleep(backoff)

			if backoff < 5*time.Second {
				backoff *= 2
			}
			continue
		}

		backoff = time.Second

		p.mu.Lock()
		old := p.conn
		p.conn = conn
		p.mu.Unlock()

		if old != nil {
			_ = old.Close()

			p.pending.Range(func(key, value any) bool {
				ch := value.(chan Response)
				ch <- Response{Error: "mpv reconnected"}
				close(ch)
				p.pending.Delete(key)
				return true
			})
		}

		go p.readerLoop(conn)
		go p.initObservations(conn)
	}
}

func (p *Player) writerLoop() {
	for {
		select {
		case req := <-p.writeCh:
			p.mu.RLock()
			conn := p.conn
			p.mu.RUnlock()

			if conn == nil {
				// fail fast, no panic
				if ch, ok := p.pending.Load(req.id); ok {
					ch.(chan Response) <- Response{Error: "mpv not connected"}
					p.pending.Delete(req.id)
				}
				continue
			}

			_, err := conn.Write(req.payload)
			if err != nil {
				if ch, ok := p.pending.Load(req.id); ok {
					ch.(chan Response) <- Response{Error: err.Error()}
					p.pending.Delete(req.id)
				}
			}

		case <-p.closeCh:
			return
		}
	}
}

func (p *Player) readerLoop(conn net.Conn) {
	defer func() {
		_ = conn.Close()
	}()

	buf := make([]byte, 0, 4096)
	tmp := make([]byte, 1024)

	for {
		n, err := conn.Read(tmp)
		if err != nil {
			return
		}

		buf = append(buf, tmp[:n]...)

		for {
			idx := bytes.IndexByte(buf, '\n')
			if idx == -1 {
				break
			}

			line := buf[:idx]
			buf = buf[idx+1:]

			var raw map[string]any
			if err := json.Unmarshal(line, &raw); err != nil {
				continue
			}

			// -------------------------
			// RESPONSE HANDLING
			// -------------------------
			if idRaw, ok := raw["request_id"]; ok {
				idFloat, ok := idRaw.(float64)
				if !ok {
					continue
				}
				id := uint64(idFloat)

				resp := Response{
					RequestID: id,
				}

				if errStr, ok := raw["error"].(string); ok {
					resp.Error = errStr
				}

				if data, ok := raw["data"]; ok {
					resp.Data = data
				}

				if chRaw, ok := p.pending.Load(id); ok {
					if ch, ok := chRaw.(chan Response); ok {
						select {
						case ch <- resp:
						default:
						}
						close(ch)
					}
					p.pending.Delete(id)
				}

				continue
			}

			// -------------------------
			// EVENT HANDLING
			// -------------------------
			event, ok := raw["event"].(string)
			if !ok {
				continue
			}

			switch event {

			case "property-change":
				name, _ := raw["name"].(string)
				data := raw["data"]
				p.handlePropertyChange(name, data)

			default:
				// non-blocking event emission (you may want to revisit this later)
				select {
				case p.events <- Event{
					Name: event,
					Data: raw,
				}:
				default:
				}
			}
		}
	}
}

func (p *Player) handlePropertyChange(name string, data any) {
	switch name {

	case "pause":
		paused, _ := data.(bool)
		p.setPlaying(!paused)

	case "time-pos":
		// optionally store playback position
		// p.setPosition(...)
	}
}

func (p *Player) initObservations(conn net.Conn) {
	obs := []string{
		`{"command": ["observe_property", 1, "pause"]}`,
		`{"command": ["observe_property", 2, "time-pos"]}`,
		`{"command": ["observe_property", 3, "duration"]}`,
	}

	for _, cmd := range obs {
		_, _ = conn.Write(append([]byte(cmd), '\n'))
	}
}

func (p *Player) setPlaying(v bool) {
	p.stateMu.Lock()
	changed := p.playing != v
	p.playing = v
	p.stateMu.Unlock()

	if changed && p.onStateChange != nil {
		p.onStateChange()
	}
}

func (p *Player) setPosition(v any) {
	f, ok := v.(float64)
	if !ok {
		return
	}

	p.posMu.Lock()
	changed := p.position != f
	p.position = f
	p.posMu.Unlock()

	if changed && p.onStateChange != nil {
		p.onStateChange()
	}
}

func (p *Player) IsPlaying() bool {
	p.stateMu.RLock()
	defer p.stateMu.RUnlock()
	return p.playing
}

func (p *Player) Send(ctx context.Context, cmd string, args ...any) (<-chan Response, error) {
	if p.closed.Load() {
		return nil, errors.New("client closed")
	}
	tr := otel.Tracer("mpv")
	_, span := tr.Start(ctx, "mpv.send")
	defer span.End()

	id := p.reqID.Add(1)

	msg := map[string]any{
		"command":    append([]any{cmd}, args...),
		"request_id": id,
	}

	b, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	respCh := make(chan Response, 1)
	p.pending.Store(id, respCh)

	select {
	case p.writeCh <- request{
		id:      id,
		payload: append(b, '\n'),
	}:
	default:
		p.pending.Delete(id)
		return nil, errors.New("write queue full")
	}

	return respCh, nil
}

func (p *Player) SendNoWait(ctx context.Context, cmd string, args ...any) error {
	_, err := p.Send(ctx, cmd, args...)
	return err
}

func (p *Player) Events() <-chan Event {
	return p.events
}

func (p *Player) Pause(ctx context.Context, state bool) error {
	cmd := "set_property"
	args := []any{"pause", state}
	_, err := p.Send(ctx, cmd, args...)
	return err
}

func (p *Player) LoadFile(ctx context.Context, path string) error {
	_, err := p.Send(ctx, "loadfile", path)
	return err
}

func (p *Player) Seek(ctx context.Context, seconds float64) error {
	_, err := p.Send(ctx, "seek", seconds, "relative")
	return err
}

func (p *Player) Close() error {
	if p.closed.Swap(true) {
		return nil
	}

	close(p.closeCh)

	p.mu.RLock()
	if p.conn != nil {
		_ = p.conn.Close()
	}
	p.mu.RUnlock()

	return nil
}
