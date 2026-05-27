/* Package mpv contains code for communicating with
* a running instance of mpv over a unix socket
 */
package mpv

import (
	"bufio"
	"encoding/json"
	"errors"
	"net"
	"sync"
	"sync/atomic"
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
	id       uint64
	payload  []byte
	respChan chan Response
}

type Player struct {
	conn    net.Conn
	writeCh chan request
	events  chan Event
	pending sync.Map // map[uint64]chan Response

	reqID   atomic.Uint64
	closeCh chan struct{}
	closed  atomic.Bool
	Playing bool
}

func New(socketpath string) (*Player, error) {
	conn, err := net.Dial("unix", socketpath)
	if err != nil {
		return nil, err
	}

	p := &Player{
		conn:    conn,
		writeCh: make(chan request, 128),
		events:  make(chan Event, 128),
		closeCh: make(chan struct{}),
		Playing: false,
	}

	go p.writerLoop()
	go p.readerLoop()

	return p, nil
}

func (p *Player) writerLoop() {
	w := bufio.NewWriter(p.conn)

	for {
		select {
		case req := <-p.writeCh:
			_, err := w.Write(req.payload)
			if err == nil {
				err = w.Flush()
			}

			if err != nil {
				// fail pending request
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

func (p *Player) readerLoop() {
	r := bufio.NewScanner(p.conn)

	for r.Scan() {
		line := r.Bytes()

		var raw map[string]any
		if err := json.Unmarshal(line, &raw); err != nil {
			continue
		}

		// response path
		if idRaw, ok := raw["request_id"]; ok {
			id := uint64(idRaw.(float64))

			resp := Response{
				RequestID: id,
			}

			if errStr, ok := raw["error"].(string); ok {
				resp.Error = errStr
			}

			resp.Data = raw["data"]

			if ch, ok := p.pending.Load(id); ok {
				ch.(chan Response) <- resp
				close(ch.(chan Response))
				p.pending.Delete(id)
			}
			continue
		}

		// event path
		if event, ok := raw["event"].(string); ok {
			p.events <- Event{
				Name: event,
				Data: raw,
			}
		}
	}
}

func (p *Player) Send(cmd string, args ...any) (<-chan Response, error) {
	if p.closed.Load() {
		return nil, errors.New("client closed")
	}

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

	p.writeCh <- request{
		id:       id,
		payload:  append(b, '\n'),
		respChan: respCh,
	}

	return respCh, nil
}

func (p *Player) SendNoWait(cmd string, args ...any) error {
	_, err := p.Send(cmd, args...)
	return err
}

func (p *Player) Events() <-chan Event {
	return p.events
}

func (p *Player) Pause(state bool) error {
	cmd := "set_property"
	args := []any{"pause", state}
	_, err := p.Send(cmd, args...)
	p.Playing = !state
	return err
}

func (p *Player) LoadFile(path string) error {
	_, err := p.Send("loadfile", path)
	return err
}

func (p *Player) Seek(seconds float64) error {
	_, err := p.Send("seek", seconds, "relative")
	return err
}

func (p *Player) Close() error {
	if p.closed.Swap(true) {
		return nil
	}

	close(p.closeCh)
	return p.conn.Close()
}
