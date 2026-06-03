/* Package model contains structs and types needed by other parts of the code */
package model

import "os/exec"

type QueueItem struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Type   string `json:"type"` // youtube | jellyfin | file | url
	Source string `json:"source"`
}

type Session struct {
	CurrentID string `json:"currentId"`
	State     string `json:"state"`
}

type WSMessage struct {
	Type  string      `json:"type"`
	Item  *QueueItem  `json:"item,omitempty"`
	Items []QueueItem `json:"items,omitempty"`
	URL   string      `json:"url,omitempty"`
	State string      `json:"state,omitempty"`
}

type PlaybackSource struct {
	Kind     string // "url" | "stream" | "process"
	URL      string
	Cmd      *exec.Cmd
	Seekable bool
	Metadata struct {
		Title    string
		Duration int64
	}
}
