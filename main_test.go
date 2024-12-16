package main

import (
	"sort"
	"testing"
	"time"
)

func TestLRU(t *testing.T) {
	c := NewCache(5)
	c.Set("A", 1, 5, 100)
	c.Set("B", 2, 15, 3)
	c.Set("C", 3, 5, 10)
	c.Set("D", 4, 1, 15)
	c.Set("E", 5, 5, 150)
	c.Get("A")

	c.SetMaxItems(5)
	testSlice(t, c.Keys(), []string{"A", "B", "C", "D", "E"})

	c.Get("C")
	c.SetMaxItems(4)
	testSlice(t, c.Keys(), []string{"A", "B", "C", "E"})

	// Make "B" the same priority with othse items.
	// We'll pick the "E" as the eviction, as it was
	// not accessed before.
	c.Set("B", 2, 5, 3)
	c.SetMaxItems(3)
	testSlice(t, c.Keys(), []string{"A", "B", "C"})

	// Touch "A" so that it won't be selected LRU candidate.
	c.Get("A")
	c.SetMaxItems(2)
	testSlice(t, c.Keys(), []string{"A", "B"})

	c.SetMaxItems(1)
	testSlice(t, c.Keys(), []string{"A"})
}

func TestMain(t *testing.T) {
	c := NewCache(5)
	c.Set("A", 1, 5, 100)
	c.Set("B", 2, 15, 3)
	c.Set("C", 3, 5, 10)
	c.Set("D", 4, 1, 15)
	c.Set("E", 5, 5, 150)
	c.Get("C")

	c.SetMaxItems(5)
	testSlice(t, c.Keys(), []string{"A", "B", "C", "D", "E"})

	time.Sleep(5 * time.Second)
	c.SetMaxItems(4)
	testSlice(t, c.Keys(), []string{"A", "C", "D", "E"})

	c.SetMaxItems(3)
	testSlice(t, c.Keys(), []string{"A", "C", "E"})

	c.SetMaxItems(2)
	testSlice(t, c.Keys(), []string{"C", "E"})

	c.SetMaxItems(1)
	testSlice(t, c.Keys(), []string{"C"})
}

func testSlice(t *testing.T, got, want []string) {
	if !compareSlice(got, want) {
		t.Fatalf("\ngot:  %v\nwant: %v", got, want)
	}
}

func compareSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	sort.Strings(a)
	sort.Strings(b)
	for i, v := range a {
		if b[i] != v {
			return false
		}
	}
	return true
}
