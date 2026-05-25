package phonesafe

import (
	"regexp"
	"sync"
)

// regexpCache provides a thread-safe cache for compiled regular expressions.
// All metadata patterns are validated at code-gen time, so compile failures
// here indicate a programmer error (bad pattern constant).
type regexpCache struct {
	mu    sync.RWMutex
	cache map[string]*regexp.Regexp
}

func newRegexpCache() *regexpCache {
	return &regexpCache{
		cache: make(map[string]*regexp.Regexp, 512),
	}
}

// get returns a compiled regexp for the pattern. It panics if the pattern
// is invalid (all metadata patterns are pre-validated at code-gen time).
func (c *regexpCache) get(pattern string) *regexp.Regexp {
	// Fast path: read lock
	c.mu.RLock()
	re, ok := c.cache[pattern]
	c.mu.RUnlock()
	if ok {
		return re
	}

	// Slow path: compile and store
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check after acquiring write lock
	if re, ok = c.cache[pattern]; ok {
		return re
	}

	re = regexp.MustCompile(pattern)
	c.cache[pattern] = re
	return re
}

// getOrError returns a compiled regexp or an error if the pattern is invalid.
func (c *regexpCache) getOrError(pattern string) (*regexp.Regexp, error) {
	// Fast path: read lock
	c.mu.RLock()
	re, ok := c.cache[pattern]
	c.mu.RUnlock()
	if ok {
		return re, nil
	}

	// Slow path: compile and store
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check after acquiring write lock
	if re, ok = c.cache[pattern]; ok {
		return re, nil
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	c.cache[pattern] = re
	return re, nil
}
