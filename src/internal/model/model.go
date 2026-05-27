/* Package model contains structs and types needed by other parts of the code */
package model

type QueueItem struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	URL   string `json:"url"`
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
