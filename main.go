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

func (c *Cache) Keys() []string {
	keys := make([]string, 0, len(c.table))
	for k := range c.table {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

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

func (c *Cache) Set(key string, value, priority, expire int) {
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

func (c *Cache) SetMaxItems(maxItems int) {
	c.maxItems = maxItems
	c.evictItems()
}

func (c *Cache) evictItems() {
	if len(c.table) <= c.maxItems {
		return
	}

	// Expired time based eviction.
	for c.expiryQ.Len() > 0 {
		got := heap.Pop(&c.expiryQ).(*Item)
		if got == nil {
			// This shouldn't happen but just in case.
			break
		}
		item, ok := c.table[got.key]
		if !ok {
			// The item is already evicted.
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

	// Priority based eviction.
	for c.priorityQ.Len() > 0 {
		got := heap.Pop(&c.priorityQ).(*Item)
		if got == nil {
			// sanity check
			break
		}
		item, ok := c.table[got.key]
		if !ok {
			// The item is already evicted.
			continue
		}
		if !got.Equal(item) {
			// Ignore the stale item in the queue.
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
