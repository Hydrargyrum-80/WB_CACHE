package main

import (
	"container/list"
	"context"
	"sync"
	"time"
)

type ICache interface {
	Cap() int
	Len() int
	Clear() // удаляет все ключи
	Add(key, value interface{})
	AddWithTTL(key, value interface{}, ttl time.Duration) // добавляет ключ со сроком жизни ttl
	Get(key interface{}) (value interface{}, ok bool)
	Remove(key interface{})
}

type Cache struct {
	cap  int
	buf  map[interface{}]*list.Element
	l    *list.List
	lock sync.RWMutex
}

type CacheElement struct {
	value  interface{}
	cancel context.CancelFunc
}

func NewCache(cap int) *Cache {
	return &Cache{cap: cap, buf: make(map[interface{}]*list.Element, cap), l: list.New()}
}

func (c *Cache) Cap() int {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.cap
}

func (c *Cache) Len() int {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.l.Len()
}

func (c *Cache) Clear() {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.buf = make(map[interface{}]*list.Element, c.cap)
	c.l = list.New()
}

func (c *Cache) Add(key, value interface{}) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if len(c.buf) == c.cap {
		removeElement := c.l.Back()
		delete(c.buf, removeElement.Value.(CacheElement).value)
		c.l.Remove(removeElement)
	}
	c.l.PushFront(CacheElement{value, nil})
	c.buf[key] = c.l.Front()
}

func (c *Cache) Get(key interface{}) (value interface{}, ok bool) {
	c.lock.RLock()
	val, status := c.buf[key]
	c.lock.RUnlock()
	if status {
		c.lock.Lock()
		defer c.lock.Unlock()
		c.l.MoveToFront(val)
		return val.Value.(CacheElement).value, status
	} else {
		return nil, ok
	}
}

func (c *Cache) AddWithTTL(key, value interface{}, ttl time.Duration) {
	expiration := time.Now().Add(ttl)
	c.lock.Lock()
	defer c.lock.Unlock()
	if len(c.buf) == c.cap {
		removeElement := c.l.Back()
		delete(c.buf, removeElement.Value.(CacheElement).value)
		c.l.Remove(removeElement)
	}
	ctx, cancel := context.WithCancel(context.Background())
	c.l.PushFront(CacheElement{value, cancel})
	c.buf[key] = c.l.Front()
	go func(t time.Time, key interface{}, c *Cache, ctx context.Context) {
		select {
		case <-time.After(ttl):
			{
				c.Remove(key)
			}
		case <-ctx.Done():
			{
				return
			}
		}
	}(expiration, key, c, ctx)
}

func (c *Cache) Remove(key interface{}) {
	c.lock.RLock()
	val, status := c.buf[key]
	c.lock.RUnlock()
	if status {
		c.lock.Lock()
		defer c.lock.Unlock()
		c.l.Remove(val)
		cancel := val.Value.(CacheElement).cancel
		if cancel != nil {
			cancel()
		}
		delete(c.buf, key)
	}
}
