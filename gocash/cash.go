package gocash

import (
	"container/list"
)

type cacheItems map[int]*list.Element

// NewCache returns a pointer to an new Cache object
func NewCache() (cache *Cache) {
	cache = &Cache{}
	cache.setup()
	return
}

type Item struct {
	Key   int
	Value int
	Elem  interface{}
}

// Cache supports various actions on an lru cache
type Cache struct {
	size  int
	lru   *list.List // Double linked list
	cache cacheItems
	Cmds  map[string]interface{}
}

func (c *Cache) setup() {
	c.lru = list.New()
	c.cache = make(cacheItems, 1000)
}

// Call takes `cmd` string and `args` ...int and calls the appropriate function
//
// A return of -101 means return wasn't set by the caller.
func (c *Cache) Call(cmd string, args ...int) (int, error) {
	var f interface{}
	switch cmd {
	case "get":
		f = c.Get
	case "add":
		f = c.Add
	case "evict":
		f = c.Evict
	case "remove":
		f = c.Remove
	}

	ret := -101
	switch t := f.(type) {
	case func():
		if err := CheckArgs(0, args); err != nil {
			return 0, err
		}
		t()
	case func(int, int):
		if err := CheckArgs(2, args); err != nil {
			return 0, err
		}
		t(args[0], args[1])
	case func(int) int:
		if err := CheckArgs(1, args); err != nil {
			return 0, err
		}
		ret = t(args[0])
	}
	return ret, nil
}

// Evict drops the oldest accessed cache item to allow freeing up memory
func (c *Cache) Evict() {
	elem := c.lru.Back()
	delete(c.cache, (*elem).Value.(Item).Key)
	c.lru.Remove(elem)
}

// Add puts `value` in the cache accessiable by `key`
func (c *Cache) Add(key int, value int) {
	elem := c.lru.PushFront(Item{Key: key, Value: value})
	c.cache[key] = elem
}

// Get returns the value in cache accessable by `key` or -1 if it doesn't exist
func (c *Cache) Get(key int) int {
	if elem, exists := c.cache[key]; exists {
		c.lru.MoveToFront(elem)
		return elem.Value.(Item).Value
	}
	// Doesn't exist so return -1
	return -1
}

// Remove returns the value for `key` in the cache after it is removed
func (c *Cache) Remove(key int) int {
	elem := c.cache[key]
	c.lru.Remove(elem)
	delete(c.cache, key)
	return elem.Value.(Item).Value
}
