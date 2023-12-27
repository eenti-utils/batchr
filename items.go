package batchr

import "time"

type batchItems[V any] map[int]*itemsHolder[V]

func (i *batchItems[V]) add(idx int, item V) {
	if holder, exists := (*i)[idx]; exists {
		holder.add(item)
	}
}

func (i *batchItems[V]) initNewItems(idx int) {
	if _, exists := (*i)[idx]; !exists {
		(*i)[idx] = newItemsHolder[V]()
	}
}

func (i *batchItems[V]) unlinkOldItems(idx int) {
	delete(*i, idx)
}

func (i *batchItems[V]) existsAndEmpty(idx int) (exists, empty bool) {
	var holder *itemsHolder[V]
	if holder, exists = (*i)[idx]; holder != nil && exists {
		empty = holder.isEmpty()
	}
	return
}

func (i *batchItems[V]) nextItems(oldIdx, newIdx int) {
	if _, exists := (*i)[oldIdx]; exists {
		i.unlinkOldItems(oldIdx)
		i.initNewItems(newIdx)
	}
}

func (i *batchItems[V]) fetchItems(idx int) (r []V) {
	if holder, exists := (*i)[idx]; exists {
		r = holder.getItems()
	}
	return
}

func (i *batchItems[V]) fetchLastUpdated(idx int) (r *time.Time) {
	if holder, exists := (*i)[idx]; exists {
		r = holder.getLastUpdated()
	}
	return
}

func (i *batchItems[V]) size(idx int) (r int) {
	if holder, exists := (*i)[idx]; exists {
		r = len(holder.items)
	}
	return
}

type itemsHolder[V any] struct {
	items       []V
	lastUpdated *time.Time
}

func newItemsHolder[V any]() *itemsHolder[V] {
	return &itemsHolder[V]{}
}

func (h *itemsHolder[V]) add(item V) {
	h.items = append(h.items, item)
	now := time.Now()
	h.lastUpdated = &now
}

func (h *itemsHolder[V]) isEmpty() bool {
	return len(h.items) == 0
}

func (h *itemsHolder[V]) getItems() (r []V) {
	if !h.isEmpty() {
		r = h.items
	}
	return
}

func (h *itemsHolder[V]) getLastUpdated() *time.Time {
	return h.lastUpdated
}
