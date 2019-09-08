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
// TODO: Clean this up
func (c *Cache) Call(cmd string, args ...string) (string, error) {
	// All cmd's map to Title cased methods on this object prefixed by Cmd
	cmd = strings.ToLower(cmd)
	vCache := reflect.ValueOf(c)
	vFunc := reflect.Indirect(vCache.MethodByName(strings.Join([]string{"Cmd", strings.Title(cmd)}, "")))
	if vFunc.Kind() != reflect.Func {
		return "", fmt.Errorf("cmd %s not found, see `help`", cmd)
	}

	tFunc := vFunc.Type()
	numNormalArgs := tFunc.NumIn()
	funcName := runtime.FuncForPC(vFunc.Pointer()).Name()
	vArgs := make([]reflect.Value, 0, len(args))
	for i, arg := range args {
		if tFunc.IsVariadic() && i >= (numNormalArgs-1) {
			variadicArg := args[numNormalArgs-1:]
			vArgs = append(vArgs, reflect.ValueOf(variadicArg))
			break
		}
		if i >= numNormalArgs {
			return "", fmt.Errorf("%s expected %d arguments, revieved %d", funcName, numNormalArgs, len(args))
		}
		argType := reflect.TypeOf(arg)
		if in := tFunc.In(i); in != argType {
			return "", fmt.Errorf("%s's arg %d expected type %s, got %v of type %s", funcName, i, tFunc.In(i), arg, argType.Name())
		}
		vArgs = append(vArgs, reflect.ValueOf(arg))
	}

	vRets := make([]reflect.Value, 0, 2)
	if tFunc.IsVariadic() {
		vRets = vFunc.CallSlice(vArgs)
	} else {
		vRets = vFunc.Call(vArgs)
	}

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

// CmdSet puts `value` in the cache accessiable by `key`
func (c *Cache) CmdSet(key string, value string) string {
	elem := c.lru.PushFront(Item{Key: key, Value: value})
	c.cache[key] = elem
	return "OK"
}

// CmdAppend extends `key` with `value`, return is length of new string
func (c *Cache) CmdAppend(key string, value string) string {
	if elem, exists := c.cache[key]; exists {
		value = strings.Join([]string{elem.Value.(Item).Value, value}, "")
	}
	ret := c.CmdSet(key, value)
	if ret == "OK" {
		return fmt.Sprintf("(integer) %d", len(value))
	}
	return ret
}

// CmdGet returns the value in cache accessable by `key` or -1 if it doesn't exist
func (c *Cache) CmdGet(key string) string {
	if elem, exists := c.cache[key]; exists {
		c.lru.MoveToFront(elem)
		return elem.Value.(Item).Value
	}
	return "(nil)" // Doesn't exist
}

// CmdDecr sets and returns the decremented value of `key`
// TODO: dedup update code
func (c *Cache) CmdDecr(key string) string {
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
	resp := c.CmdSet(key, strconv.Itoa(value))
	if resp != "OK" {
		return resp
	}
	return strconv.Itoa(value)
}

// CmdIncr sets and returns the decremented value of `key`
// TODO: dedup update code
func (c *Cache) CmdIncr(key string) string {
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
	resp := c.CmdSet(key, strconv.Itoa(value))
	if resp != "OK" {
		return resp
	}
	return strconv.Itoa(value)
}

// CmdEcho returns arg as it was passed
func (c *Cache) CmdEcho(arg string) string {
	return arg
}

// CmdExists returns 1 if key exists, 0 otherwise
func (c *Cache) CmdExists(key string) string {
	if _, exists := c.cache[key]; exists {
		return "(integer) 1"
	}
	return "(integer) 0"
}

// CmdDel returns the value for `key` in the cache after it is removed
func (c *Cache) CmdDel(keys ...string) string {
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
