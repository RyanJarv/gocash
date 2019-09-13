package gocash

import (
	"fmt"
	"reflect"
	"runtime"
	"strconv"
	"strings"
)

// ValueType is redis's 1byte declaration of encoding
type ValueType uint8

// Declares ValueType Encodings
const (
	String ValueType = iota // 0
	List
	Set
	SortedSet
	HashEnc
	_
	_
	_
	_
	ZipMap // 9
	ZipList
	IntSet
	SortedSetInZipList
	HashMapInZipList
)

// Similar to redis's RDB format except key and value are not stored the same way
type redisElem struct {
	valueType ValueType
	value     interface{}
}

// NewCache returns a pointer to an new Cache object
func NewCache() (cache *Cache) {
	return &Cache{cache: NewHash()}
}

// Cache supports various actions on an lru cache
type Cache struct {
	cache *Hash
	Cmds  map[string]interface{}
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

// CmdSet puts `value` in the cache accessiable by `key`
func (c *Cache) CmdSet(key, value string) string {
	elem := redisElem{
		valueType: String,
		value:     value,
	}
	if c.cache.Put(key, elem) != nil {
		panic(fmt.Errorf("Unable to set key %v to value %v", key, value))
	}
	return "OK"
}

// CmdAppend extends `key` with `value`, return is length of new string
func (c *Cache) CmdAppend(key, value string) string {
	if tmpValue, err := c.cache.Get(key); err == nil {
		if tmpValue.(redisElem).valueType != String {
			panic(fmt.Errorf("Key %v is not of type String", key))
		}
		value = strings.Join([]string{tmpValue.(redisElem).value.(string), value}, "")
	}
	ret := c.CmdSet(key, value)
	if string(ret) == "OK" {
		return fmt.Sprintf("(integer) %d", len(value))
	}
	return ret
}

// CmdGet returns the value in cache accessable by `key` or -1 if it doesn't exist
func (c *Cache) CmdGet(key string) string {
	if value, err := c.cache.Get(key); err == nil {
		elem := value.(redisElem)
		switch elem.valueType {
		case String:
			return elem.value.(string)
		default:
			panic(fmt.Sprintf("CmdGet not implemented for the value type of %v", elem.value))
		}
	}
	return "(nil)" // Doesn't exist
}

// CmdDecr sets and returns the decremented value of `key`
// TODO: dedup update code
func (c *Cache) CmdDecr(key string) string {
	var err error
	var value interface{}
	var elem redisElem
	var tmpValue int
	if value, err = c.cache.Get(key); err != nil {
		elem = value.(redisElem)
		if elem.valueType != String {
			panic(fmt.Errorf("CmdDecr not implented for value type of value %s", elem.value))
		}
		tmpValue, err = strconv.Atoi(elem.value.(string))
		if err != nil {
			return "Err value is not an integer or is out of range"
		}
		tmpValue--
	}
	resp := c.CmdSet(key, strconv.Itoa(tmpValue))
	if string(resp) != "OK" {
		return resp
	}
	return strconv.Itoa(tmpValue)
}

// CmdIncr sets and returns the decremented value of `key`
// TODO: dedup update code
func (c *Cache) CmdIncr(key string) string {
	var err error
	var value interface{}
	var elem redisElem
	var tmpValue int
	if value, err = c.cache.Get(key); err != nil {
		elem = value.(redisElem)
		if elem.valueType != String {
			panic(fmt.Errorf("CmdIncr not implented for value type of value %s", elem.value))
		}
		tmpValue, err = strconv.Atoi(elem.value.(string))
		if err != nil {
			return "Err value is not an integer or is out of range"
		}
		tmpValue++
	}
	resp := c.CmdSet(key, strconv.Itoa(tmpValue))
	if string(resp) != "OK" {
		return resp
	}
	return strconv.Itoa(tmpValue)
}

// CmdEcho returns arg as it was passed
func (c *Cache) CmdEcho(arg string) string {
	return arg
}

// CmdExists returns 1 if key exists, 0 otherwise
func (c *Cache) CmdExists(key string) string {
	if _, err := c.cache.Get(key); err != nil {
		return "(integer) 1"
	}
	return "(integer) 0"
}

// CmdDel returns the value for `key` in the cache after it is removed
func (c *Cache) CmdDel(keys ...string) string {
	deleted := 0
	for _, key := range keys {
		_, err := c.cache.Remove(key)
		if err != nil {
			deleted++
		}
	}
	return fmt.Sprintf("(integer) %d", deleted)
}

// CmdSave snapshots the current cache
func (c *Cache) CmdSave() string {
	c.cache.Snapshot()
	return "OK"
}
