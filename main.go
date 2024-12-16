// A LRU cache in Golang.
package main

import (
	"container/heap"
	"fmt"
	"sort"
	"time"
)

// Cache is a priority expiry LRU cache.
type Cache struct {
	maxItems  int
	table     map[string]*Item
	priorityQ PriorityQueue
	expiryQ   ExpiryQueue
}

// Keys returns the keys in the cache in sorted order.
func (c *Cache) Keys() []string {
	keys := make([]string, 0, len(c.table))
	for k := range c.table {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// NewCache creates a new priority expiry LRU cache.
func NewCache(maxItems int) *Cache {
	var priorityQ PriorityQueue
	heap.Init(&priorityQ)
	var expiryQ ExpiryQueue
	heap.Init(&expiryQ)
	return &Cache{
		maxItems:  maxItems,
		table:     make(map[string]*Item),
		priorityQ: priorityQ,
		expiryQ:   expiryQ,
	}
}

// Get returns the value for the key or -1 if:
//
// 1. the key does not exists in the cache
// 2. the item is already expired
func (c *Cache) Get(key string) int {
	item, ok := c.table[key]
	if ok {
		if time.Now().Before(item.expire) {
			item.access = time.Now()
			return item.value
		}
		delete(c.table, key)
	}
	return -1
}

// Set sets the new value for the key, with priority and expire time in seconds.
func (c *Cache) Set(key string, value, priority, expire int) {
	// Create a brand new item and insert it to all three field,
	// table, priorityQ, expiryQ.  This way, we don't need to
	// call the heap.Fix operation, which is O(log n) runtime.
	expireDuration := time.Duration(expire) * time.Second
	item := Item{
		key:      key,
		value:    value,
		priority: priority,
		expire:   time.Now().Add(expireDuration),
		access:   time.Now(),
	}
	c.table[key] = &item
	heap.Push(&c.priorityQ, &item)
	heap.Push(&c.expiryQ, &item)
	c.evictItems()
}

// SetMaxItems update the maximum value.
//
// It evicts items in case the actual cache size is more than
// the maximum value.
func (c *Cache) SetMaxItems(maxItems int) {
	c.maxItems = maxItems
	c.evictItems()
}

// evictItems will evict items from the cache to make room for new ones.
func (c *Cache) evictItems() {
	// Cache has a capacity, do nothing.
	if len(c.table) <= c.maxItems {
		return
	}

	// Evicts expired items first, if any.
	for c.expiryQ.Len() > 0 {
		got := heap.Pop(&c.expiryQ).(*Item)
		if got == nil {
			// This shouldn't happen but just in case.
			break
		}
		item, ok := c.table[got.key]
		if !ok {
			// The item is already evicted based on the
			// priority.
			continue
		}
		if !got.Equal(item) {
			// Ignore the stale item in the queue.
			continue
		}
		if item.expire.After(time.Now()) {
			// No more expired items, try the priority based
			// eviction next.
			heap.Push(&c.expiryQ, item)
			break
		}
		delete(c.table, item.key)
		if len(c.table) <= c.maxItems {
			// done.
			return
		}
	}

	// Evicts items based on the priority.
	//
	// Evicts LRU, Least Recent Updated, items in case of the same
	// priority.
	for c.priorityQ.Len() > 0 {
		got := heap.Pop(&c.priorityQ).(*Item)
		if got == nil {
			// This shouldn't happen but, just sanity check
			break
		}
		item, ok := c.table[got.key]
		if !ok {
			// The item is already evicted based on the
			// expiration time.
			continue
		}
		if !got.Equal(item) {
			// The item had been updated by Set() API.
			// ignore the item in the queue.
			continue
		}
		delete(c.table, item.key)
		if len(c.table) <= c.maxItems {
			// done.
			return
		}
	}
}

// Item holds the value and metadata.
type Item struct {
	key      string
	value    int
	priority int
	access   time.Time
	expire   time.Time
}

// Equal checks the item value and the metadata except the access time.
func (i *Item) Equal(j *Item) bool {
	if i.key != j.key {
		return false
	}
	if i.value != j.value {
		return false
	}
	if i.priority != j.priority {
		return false
	}
	if i.expire != j.expire {
		return false
	}
	return true
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
	if pq.BaseQueue[i].priority == pq.BaseQueue[j].priority {
		// Use the LRU logic for the same priority items.
		return pq.BaseQueue[i].access.Before(pq.BaseQueue[j].access)
	} else {
		return pq.BaseQueue[i].priority < pq.BaseQueue[j].priority
	}
}

func main() {
	c := NewCache(5)
	c.Set("A", 1, 5, 100)
	c.Set("B", 2, 15, 3)
	c.Set("C", 3, 5, 10)
	c.Set("D", 4, 1, 15)
	c.Set("E", 5, 5, 150)
	c.Get("C")

	c.SetMaxItems(5)
	fmt.Println(c.Keys())

	time.Sleep(5 * time.Second)
	c.SetMaxItems(4)
	fmt.Println(c.Keys())

	c.SetMaxItems(3)
	fmt.Println(c.Keys())

	c.SetMaxItems(2)
	fmt.Println(c.Keys())

	c.SetMaxItems(1)
	fmt.Println(c.Keys())
}
