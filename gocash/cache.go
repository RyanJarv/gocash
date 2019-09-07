package gocash

import (
	"container/list"
	"fmt"
	"reflect"
	"runtime"
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

	resp, err := c.call(f, args...)
	if err != nil {
		return "", err
	}
	return resp, nil
}

func (c *Cache) call(f interface{}, args ...string) (string, error) {
	rf := reflect.TypeOf(f)
	vf := reflect.ValueOf(f)
	if rf.Kind() != reflect.Func {
		panic("expects a function")
	}

	funcName := runtime.FuncForPC(vf.Pointer()).Name()
	for i, arg := range args {
		if expected := rf.NumIn(); i >= expected {
			return "", fmt.Errorf("%s expected %d arguments, revieved %d", funcName, expected, len(args))
		}
		argType := reflect.TypeOf(arg)
		if rf.In(i) != argType {
			return "", fmt.Errorf("%s's arg %d expected type %s, got %v of type %s", funcName, i, rf.In(i), arg, argType.Name())
		}
	}

	vArgs := make([]reflect.Value, 0, len(args))
	for _, arg := range args {
		vArgs = append(vArgs, reflect.ValueOf(arg))
	}

	vRets := vf.Call(vArgs)
	if vRets[0].Type().Kind() != reflect.String {
		return "", fmt.Errorf("Expected string for %s first return type", funcName)
	}
	return vRets[0].String(), nil
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
