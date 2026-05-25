package queue

import (
	"media-jukebox-backend/internal/model"
	"sync"
)

type Queue struct {
	mu    sync.Mutex
	items []model.QueueItem
	idx   int
}

func New() *Queue {
	return &Queue{
		items: []model.QueueItem{},
		idx:   -1,
	}
}

func (q *Queue) Add(item model.QueueItem) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.items = append(q.items, item)
}

func (q *Queue) Items() []model.QueueItem {
	q.mu.Lock()
	defer q.mu.Unlock()

	copyItems := make([]model.QueueItem, len(q.items))
	copy(copyItems, q.items)

	return copyItems
}

func (q *Queue) Current() *model.QueueItem {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.idx < 0 || q.idx >= len(q.items) {
		return nil
	}

	item := q.items[q.idx]
	return &item
}

func (q *Queue) Start() *model.QueueItem {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.items) == 0 {
		return nil
	}

	if q.idx == -1 {
		q.idx = 0
	}

	item := q.items[q.idx]
	return &item
}

func (q *Queue) Next() *model.QueueItem {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.items) == 0 {
		return nil
	}

	if q.idx+1 >= len(q.items) {
		q.idx = len(q.items)
		return nil
	}

	q.idx++

	item := q.items[q.idx]
	return &item
}

func (q *Queue) Prev() *model.QueueItem {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.items) == 0 {
		return nil
	}

	if q.idx <= 0 {
		q.idx = 0
		return nil
	}

	q.idx--

	item := q.items[q.idx]
	return &item
}
