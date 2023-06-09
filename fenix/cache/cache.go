package cache

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/gomodule/redigo/redis"
)

type Cache interface {
	Exists(string) (bool, error)
	Get(string) (interface{}, error)
	Set(string, interface{}, ...int) error
	Remove(string) error
	EmptyByMatch(string) error
	Empty() error
}

type RedisCache struct {
	Conn   *redis.Pool
	Prefix string // used to prefix keys with something unique in case they share the same ids
}

type Entry map[string]interface{}

func (f *RedisCache) Exists(str string) (bool, error) {
	key := fmt.Sprintf("%s:%s", f.Prefix, str)
	conn := f.Conn.Get()
	defer conn.Close()

	ok, err := redis.Bool(conn.Do("EXISTS", key))
	if err != nil {
		return false, err
	}

	return ok, nil
}

func encode(item Entry) ([]byte, error) {
	b := bytes.Buffer{}
	e := gob.NewEncoder(&b)
	err := e.Encode(item)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func decode(str string) (Entry, error) {
	item := Entry{}
	b := bytes.Buffer{}
	b.Write([]byte(str))
	d := gob.NewDecoder(&b)
	err := d.Decode(&item)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (f *RedisCache) Get(str string) (interface{}, error) {
	key := fmt.Sprintf("%s:%s", f.Prefix, str)
	conn := f.Conn.Get()
	defer conn.Close()

	cacheEntry, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		return nil, err
	}

	decoded, err := decode(string(cacheEntry))
	if err != nil {
		return nil, err
	}

	item := decoded[key]

	return item, nil
}

func (f *RedisCache) Set(str string, val interface{}, expiry ...int) error {
	key := fmt.Sprintf("%s:%s", f.Prefix, str)
	conn := f.Conn.Get()
	defer conn.Close()

	entry := Entry{}
	entry[key] = val
	encoded, err := encode(entry)
	if err != nil {
		return err
	}

	if len(expiry) > 0 {
		_, err := conn.Do("SETEX", key, expiry[0], string(encoded))
		if err != nil {
			return err
		}
	} else {
		_, err := conn.Do("SETEX", key, string(encoded))
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *RedisCache) Remove(str string) error {
	key := fmt.Sprintf("%s:%s", f.Prefix, str)
	conn := f.Conn.Get()
	defer conn.Close()

	_, err := conn.Do("DEL", key)
	if err != nil {
		return err
	}
	return nil
}

func (f *RedisCache) EmptyByMatch(str string) error {
	key := fmt.Sprintf("%s:%s", f.Prefix, str)
	conn := f.Conn.Get()
	defer conn.Close()

	keys, err := f.getKeys(key)
	if err != nil {
		return err
	}

	for _, v := range keys {
		err := f.Remove(v)
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *RedisCache) Empty() error {
	key := fmt.Sprintf("%s:", f.Prefix)
	conn := f.Conn.Get()
	defer conn.Close()

	keys, err := f.getKeys(key)
	if err != nil {
		return err
	}

	for _, v := range keys {
		err = f.Remove(v)
		if err != nil {
			return err
		}
	}

	return nil
}

func (f *RedisCache) getKeys(pattern string) ([]string, error) {
	conn := f.Conn.Get()
	defer conn.Close()

	iter := 0
	keys := []string{}

	for {
		arr, err := redis.Values(conn.Do("SCAN", iter, "MATCH", fmt.Sprintf("%s*", pattern)))
		if err != nil {
			return keys, err
		}

		iter, _ := redis.Int(arr[0], nil)
		k, _ := redis.Strings(arr[1], nil)
		keys = append(keys, k...)

		if iter == 0 {
			break
		}
	}

	return keys, nil
}
