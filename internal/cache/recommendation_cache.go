package cache

import (
	"sync"
	"time"
)

type entry struct {
	value     interface{}
	expiresAt time.Time
}

var (
	store sync.Map
	once  sync.Once
)

func init() {
	once.Do(func() {
		go func() {
			ticker := time.NewTicker(time.Minute)
			defer ticker.Stop()
			for range ticker.C {
				store.Range(func(k, v interface{}) bool {
					if e, ok := v.(entry); ok && time.Now().After(e.expiresAt) {
						store.Delete(k)
					}
					return true
				})
			}
		}()
	})
}

// Set stores value under key with the given TTL.
func Set(key string, value interface{}, ttl time.Duration) {
	store.Store(key, entry{value: value, expiresAt: time.Now().Add(ttl)})
}

// Get retrieves a cached value. Returns (value, true) on hit, (nil, false) on miss or expiry.
func Get(key string) (interface{}, bool) {
	raw, ok := store.Load(key)
	if !ok {
		return nil, false
	}
	e := raw.(entry)
	if time.Now().After(e.expiresAt) {
		store.Delete(key)
		return nil, false
	}
	return e.value, true
}
