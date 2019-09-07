package gocash

import (
	"container/list"
	"fmt"
	"strconv"
	"strings"
)

type cacheItems map[string]*list.Element

// NewCache returns a pointer to an new Cache object
func NewCache() (cache *Cache) {
	cache = &Cache{}
	cache.setup()
	return
}

type Item struct {
	Key   string
	Value string
	Elem  interface{}
}

// Cache supports various actions on an lru cache
type Cache struct {
	lru   *list.List // Double linked list
	cache cacheItems
	Cmds  map[string]interface{}
}

func (c *Cache) setup() {
	c.lru = list.New()
	c.cache = make(cacheItems, 1000)
}

// Call takes `cmd` string and `args` ...int and calls the appropriate function
func (c *Cache) Call(cmd string, args ...string) (string, error) {
	var f interface{}
	switch cmd {
	case "get":
		f = c.Get
	case "exists":
		f = c.Exists
	case "set":
		f = c.Set
	case "decr":
		f = c.Decr
	case "incr":
		f = c.Incr
	case "append":
		f = c.Append
	case "evict":
		f = c.Evict
	case "del":
		f = c.Delete
	case "echo":
		f = c.Echo
	case "help":
		return "Only get, exists, set, apppend, echo, evict (non-standard), and del are supported right now", nil
	default:
		return "", fmt.Errorf("Unkown or disabled command '%s'", cmd)
	}

	var ret string
	switch t := f.(type) {
	case func() string:
		if err := CheckArgs(0, args); err != nil {
			return "", err
		}
		t()
	case func(string, string) string:
		if err := CheckArgs(2, args); err != nil {
			return "", err
		}
		t(args[0], args[1])
	case func(string) string:
		if err := CheckArgs(1, args); err != nil {
			return "", err
		}
		ret = t(args[0])
	case func(...string) string:
		ret = t(args[0])
	default:
		return "", fmt.Errorf("Cmd signature not found for %v", t)
	}
	return ret, nil
}

// Evict drops the oldest accessed cache item to allow freeing up memory
func (c *Cache) Evict() string {
	elem := c.lru.Back()
	delete(c.cache, (*elem).Value.(Item).Key)
	c.lru.Remove(elem)
	return "OK"
}

// Set puts `value` in the cache accessiable by `key`
func (c *Cache) Set(key string, value string) string {
	elem := c.lru.PushFront(Item{Key: key, Value: value})
	c.cache[key] = elem
	return "OK"
}

// Append extends `key` with `value`, return is length of new string
func (c *Cache) Append(key string, value string) string {
	if elem, exists := c.cache[key]; exists {
		value = strings.Join([]string{elem.Value.(Item).Value, value}, "")
	}
	ret := c.Set(key, value)
	if ret == "OK" {
		return fmt.Sprintf("(integer) %d", len(value))
	}
	return ret
}

// Get returns the value in cache accessable by `key` or -1 if it doesn't exist
func (c *Cache) Get(key string) string {
	if elem, exists := c.cache[key]; exists {
		c.lru.MoveToFront(elem)
		return elem.Value.(Item).Value
	}
	return "(nil)" // Doesn't exist
}

// Decr sets and returns the decremented value of `key`
// TODO: dedup update code
func (c *Cache) Decr(key string) string {
	var err error
	var value int
	if elem, exists := c.cache[key]; exists {
		c.lru.MoveToFront(elem)
		value, err = strconv.Atoi(elem.Value.(Item).Value)
		if err != nil {
			return "Err value is not an integer or is out of range"
		}
		value--
	}
	resp := c.Set(key, strconv.Itoa(value))
	if resp != "OK" {
		return resp
	}
	return strconv.Itoa(value)
}

// Incr sets and returns the decremented value of `key`
// TODO: dedup update code
func (c *Cache) Incr(key string) string {
	var err error
	var value int
	if elem, exists := c.cache[key]; exists {
		c.lru.MoveToFront(elem)
		value, err = strconv.Atoi(elem.Value.(Item).Value)
		if err != nil {
			return "Err value is not an integer or is out of range"
		}
		value++
	}
	resp := c.Set(key, strconv.Itoa(value))
	if resp != "OK" {
		return resp
	}
	return strconv.Itoa(value)
}

// Echo returns arg as it was passed
func (c *Cache) Echo(arg string) string {
	return arg
}

// Exists returns 1 if key exists, 0 otherwise
func (c *Cache) Exists(key string) string {
	if _, exists := c.cache[key]; exists {
		return "(integer) 1"
	}
	return "(integer) 0"
}

// Delete returns the value for `key` in the cache after it is removed
func (c *Cache) Delete(keys ...string) string {
	deleted := 0
	for _, key := range keys {
		elem, found := c.cache[key]
		if found {
			deleted++
		}
		c.lru.Remove(elem)
		delete(c.cache, key)
	}
	return fmt.Sprintf("(integer) %d", deleted)
}
