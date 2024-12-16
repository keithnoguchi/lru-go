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
			// Priority queue could be affected by the
			// updated access time for the LRU operation.
			heap.Fix(&c.priorityQ, item.priorityIndex)
			return item.value
		}
		delete(c.table, key)
	}
	return -1
}

// Set sets the new value for the key, with priority and expire time in seconds.
func (c *Cache) Set(key string, value, priority, expire int) {
	accessTime := time.Now()
	expireTime := accessTime.Add(time.Duration(expire) * time.Second)
	if item := c.table[key]; item != nil {
		item.value = value
		item.priority = priority
		item.access = accessTime
		heap.Fix(&c.priorityQ, item.priorityIndex)
		if item.expire != expireTime {
			item.expire = expireTime
			heap.Fix(&c.expiryQ, item.expiryIndex)
		}
	} else {
		item := Item{
			key:      key,
			value:    value,
			priority: priority,
			expire:   expireTime,
			access:   accessTime,
		}
		c.table[key] = &item
		heap.Push(&c.priorityQ, &item)
		heap.Push(&c.expiryQ, &item)
	}
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
		// Peak the candidate eviction item.
		got := c.expiryQ[0]
		if got == nil {
			// This shouldn't happen but just in case.
			break
		}
		item, ok := c.table[got.key]
		if !ok {
			// The item is already evicted.  This could be
			// through the proirity based eviction.
			//
			// Remove it from the expiry queue and try the
			// next one.
			heap.Pop(&c.expiryQ)
			continue
		}
		if item.expire.After(time.Now()) {
			// No more expired items, try the priority based
			// eviction next.
			break
		}
		// Evict the item.
		heap.Pop(&c.expiryQ)
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
		item, ok := c.table[got.key]
		if !ok {
			// The item is already evicted.  This could be
			// through the expiry based eviction.
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
	key           string
	value         int
	priority      int
	access        time.Time
	expire        time.Time
	priorityIndex int
	expiryIndex   int
}

// https://pkg.go.dev/container/heap#example-package-PriorityQueue
type ExpiryQueue []*Item

func (pq ExpiryQueue) Len() int { return len(pq) }
func (pq ExpiryQueue) Less(i, j int) bool {
	// To pop the oldest expire time item first.
	return pq[i].expire.Before(pq[j].expire)
}
func (pq ExpiryQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].expiryIndex = i
	pq[j].expiryIndex = j
}
func (pq *ExpiryQueue) Push(x any) {
	n := len(*pq)
	item := x.(*Item)
	item.expiryIndex = n
	*pq = append(*pq, item)
}
func (pq *ExpiryQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.expiryIndex = -1
	*pq = old[0 : n-1]
	return item
}

// Priority with LRU based priority queue.
type PriorityQueue []*Item

func (pq PriorityQueue) Len() int { return len(pq) }
func (pq PriorityQueue) Less(i, j int) bool {
	// To pop the lowest priority item first.
	if pq[i].priority == pq[j].priority {
		// Pick the LRU item.
		return pq[i].access.Before(pq[j].access)
	} else {
		return pq[i].priority < pq[j].priority
	}
}
func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].priorityIndex = i
	pq[j].priorityIndex = j
}
func (pq *PriorityQueue) Push(x any) {
	n := len(*pq)
	item := x.(*Item)
	item.priorityIndex = n
	*pq = append(*pq, item)
}
func (pq *PriorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.priorityIndex = -1
	*pq = old[0 : n-1]
	return item
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
	fmt.Println(c.Keys()) // [A B C D E]

	time.Sleep(5 * time.Second)
	c.SetMaxItems(4)
	fmt.Println(c.Keys()) // [A C D E]

	c.SetMaxItems(3)
	fmt.Println(c.Keys()) // [A C E]

	c.SetMaxItems(2)
	fmt.Println(c.Keys()) // [C E]

	c.SetMaxItems(1)
	fmt.Println(c.Keys()) // [C]
}
