// Package cachelruttl is a naive implementation of a cache to use with awssecretmanager.
//
// LRU: the size is fixed, it starts removing old entries when full.
//
// TTL: it will not return old entries. Time precision: second.
// Old entries will still be in the cache, but filtered out on the Get().
package cachelruttl

import (
	"time"

	"github.com/golang/groupcache/lru"
)

type (
	Cache struct {
		subCache *lru.Cache
		ttl      time.Duration
	}

	entry struct {
		value   interface{}
		expired int64 //could be time.Time, but we don't realistically need a better precision than seconds.
	}
)

func New(size int, ttl time.Duration) *Cache {
	return &Cache{
		subCache: lru.New(size),
		ttl:      ttl,
	}
}

func (c Cache) addWithExpired(key, value interface{}, expired int64) {
	c.subCache.Add(key, entry{
		value:   value,
		expired: expired,
	})
}

func (c Cache) Add(key, value interface{}) {
	c.addWithExpired(key, value, time.Now().Add(c.ttl).Unix())
}

func (c Cache) Get(key interface{}) (value interface{}, ok bool) {
	value, ok = c.subCache.Get(key)
	if !ok {
		return nil, false
	}
	res := value.(entry)
	if time.Now().Unix() > res.expired {
		//not even deleting the entry. Filling all the cache
		return nil, false
	}
	return res.value, ok
}
