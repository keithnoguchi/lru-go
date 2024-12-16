// A LRU cache in Golang.
package main

import (
	_ "container/heap"
	"time"
)

// Cache is a priority expiry LRU cache.
type Cache struct {
	maxItems  int
	table     map[string]*Item
	expiryQ   expiryQueue
	priorityQ priorityQueue
}

func NewCache(maxItems int) *Cache {
	var expiryQ expiryQueue
	var priorityQ priorityQueue
	//heap.Init(&expiryQ)
	//heap.Init(&priorityQ)
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

type expiryQueue []*Item
type priorityQueue []*Item

func main() {
	c := NewCache(5)
	c.SetMaxItems(5)
}
