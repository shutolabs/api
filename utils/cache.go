package utils

import (
	"fmt"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
)

type CacheEntry[T any] struct {
    Value      T
    CreatedAt  time.Time
    ExpiresAt  time.Time
    StaleAt    time.Time    // New field for SWR
    IsStale    bool         // Flag to track stale state
}

type Cache[T any] struct {
    cache *lru.Cache[string, *CacheEntry[T]]
    mu    sync.Mutex // Mutex for background updates
}

type CacheOptions struct {
    MaxSize int
}

type GetCachedOptions struct {
    Key           string
    TTL           time.Duration
    StaleTime     time.Duration    // Time until data becomes stale
    GetFreshValue func() (interface{}, error)
}

func NewCache[T any](opts CacheOptions) (*Cache[T], error) {
    c, err := lru.New[string, *CacheEntry[T]](opts.MaxSize)
    if err != nil {
        return nil, err
    }
    return &Cache[T]{cache: c}, nil
}

func (c *Cache[T]) GetCached(opts GetCachedOptions) (T, error) {
    var empty T
    now := time.Now()

    // Check if we have a cached entry
    if entry, ok := c.cache.Get(opts.Key); ok {
        // If the data is not expired, return it
        if now.Before(entry.ExpiresAt) {
            return entry.Value, nil
        }

        // If the data is stale but not completely expired, return it and refresh in background
        if !entry.IsStale && now.Before(entry.StaleAt) {
            go c.refreshInBackground(opts)
            return entry.Value, nil
        }
    }

    // No valid cache entry, get fresh value
    value, err := c.getFreshValue(opts)
    if err != nil {
        return empty, err
    }

    return value, nil
}

func (c *Cache[T]) refreshInBackground(opts GetCachedOptions) {
    c.mu.Lock()
    defer c.mu.Unlock()

    // Double-check if someone else already updated the cache
    if entry, ok := c.cache.Get(opts.Key); ok {
        if !entry.IsStale {
            return
        }
    }

    // Mark as stale while updating
    if entry, ok := c.cache.Get(opts.Key); ok {
        entry.IsStale = true
        c.cache.Add(opts.Key, entry)
    }

    // Get fresh value
    _, _ = c.getFreshValue(opts) // Ignore errors in background refresh
}

func (c *Cache[T]) getFreshValue(opts GetCachedOptions) (T, error) {
    var empty T

    value, err := opts.GetFreshValue()
    if err != nil {
        return empty, err
    }

    typedValue, ok := value.(T)
    if !ok {
        return empty, fmt.Errorf("value type assertion failed")
    }

    now := time.Now()
    entry := &CacheEntry[T]{
        Value:     typedValue,
        CreatedAt: now,
        ExpiresAt: now.Add(opts.TTL),
        StaleAt:   now.Add(opts.StaleTime),
        IsStale:   false,
    }
    c.cache.Add(opts.Key, entry)

    return typedValue, nil
} 