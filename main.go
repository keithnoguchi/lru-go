// A LRU cache in Golang.
package main

import (
	"container/heap"
	"time"
)

// Cache is a priority expiry LRU cache.
type Cache struct {
	maxItems  int
	table     map[string]*Item
	priorityQ PriorityQueue
	expiryQ   ExpiryQueue
}

func NewCache(maxItems int) *Cache {
	var priorityQ PriorityQueue
	heap.Init(&priorityQ)
	var expiryQ ExpiryQueue
	heap.Init(&expiryQ)
	return &Cache{
		maxItems:  maxItems,
		table:     make(map[string]*Item),
		expiryQ:   expiryQ,
		priorityQ: priorityQ,
	}
}

// Item holds the value with metadata.
type Item struct {
	key      string
	value    int
	priority int
	access   time.Time
	expire   time.Time
}

func (c *Cache) Get(key string) int {
	item, ok := c.table[key]
	if ok {
		if item.expire.After(time.Now()) {
			return item.value
		}
		delete(c.table, key)
	}
	return -1
}

func (c *Cache) Set(key string, value, priority, expire int) {
	duration := time.Duration(expire) * time.Second
	item := Item{
		key:      key,
		value:    value,
		priority: priority,
		expire:   time.Now().Add(duration),
		access:   time.Now(),
	}
	c.table[key] = &item
	c.evictItems()
}

func (c *Cache) SetMaxItems(maxItems int) {
	c.maxItems = maxItems
	c.evictItems()
}

func (c *Cache) evictItems() {
}

// https://pkg.go.dev/container/heap#example-package-PriorityQueue
type BaseQueue []*Item
type ExpiryQueue struct {
	BaseQueue
}
type PriorityQueue struct {
	BaseQueue
}

func (pq BaseQueue) Len() int { return len(pq) }
func (pq BaseQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}
func (pq *BaseQueue) Push(x any) {
	item := x.(*Item)
	*pq = append(*pq, item)
}
func (pq *BaseQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	*pq = old[0 : n-1]
	return item
}
func (pq ExpiryQueue) Less(i, j int) bool {
	// To pop the oldest expire time item first.
	return pq.BaseQueue[i].expire.Before(pq.BaseQueue[j].expire)
}
func (pq PriorityQueue) Less(i, j int) bool {
	// To pop the lowest priority item first.
	return pq.BaseQueue[i].priority < pq.BaseQueue[j].priority
}

func main() {
	c := NewCache(5)
	c.SetMaxItems(5)
}
