package cache

import (
	"sync"
	"time"
)

// Item represents a cache item
type Item struct {
	Value      interface{}
	Expiration int64
}

// Expired returns whether the item has expired
func (item Item) Expired() bool {
	if item.Expiration == 0 {
		return false
	}
	return time.Now().UnixNano() > item.Expiration
}

// Cache is a simple in-memory cache with expiration
type Cache struct {
	items             map[string]Item
	mu                sync.RWMutex
	defaultExpiration time.Duration
	cleanupInterval   time.Duration
	stopCleanup       chan bool
}

// NewCache creates a new cache with the given default expiration and cleanup interval
func NewCache(defaultExpiration, cleanupInterval time.Duration) *Cache {
	cache := &Cache{
		items:             make(map[string]Item),
		defaultExpiration: defaultExpiration,
		cleanupInterval:   cleanupInterval,
		stopCleanup:       make(chan bool),
	}
	
	// Start the cleanup goroutine
	if cleanupInterval > 0 {
		go cache.startCleanupTimer()
	}
	
	return cache
}

// startCleanupTimer starts the cleanup timer
func (c *Cache) startCleanupTimer() {
	ticker := time.NewTicker(c.cleanupInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			c.DeleteExpired()
		case <-c.stopCleanup:
			return
		}
	}
}

// Set adds an item to the cache with the default expiration
func (c *Cache) Set(key string, value interface{}) {
	c.SetWithExpiration(key, value, c.defaultExpiration)
}

// SetWithExpiration adds an item to the cache with a specific expiration
func (c *Cache) SetWithExpiration(key string, value interface{}, d time.Duration) {
	var expiration int64
	
	if d == 0 {
		// 0 means use default expiration
		d = c.defaultExpiration
	}
	
	if d > 0 {
		expiration = time.Now().Add(d).UnixNano()
	}
	
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.items[key] = Item{
		Value:      value,
		Expiration: expiration,
	}
}

// Get gets an item from the cache
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	item, found := c.items[key]
	if !found {
		return nil, false
	}
	
	// Check if the item has expired
	if item.Expired() {
		return nil, false
	}
	
	return item.Value, true
}

// Delete deletes an item from the cache
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	delete(c.items, key)
}

// DeleteExpired deletes all expired items from the cache
func (c *Cache) DeleteExpired() {
	now := time.Now().UnixNano()
	
	c.mu.Lock()
	defer c.mu.Unlock()
	
	for k, v := range c.items {
		if v.Expiration > 0 && now > v.Expiration {
			delete(c.items, k)
		}
	}
}

// Flush deletes all items from the cache
func (c *Cache) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.items = make(map[string]Item)
}

// Count returns the number of items in the cache
func (c *Cache) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	return len(c.items)
}

// Shutdown stops the cleanup goroutine
func (c *Cache) Shutdown() {
	if c.cleanupInterval > 0 {
		c.stopCleanup <- true
	}
}

// LRUCache implements a Least Recently Used (LRU) cache
type LRUCache struct {
	capacity          int
	items             map[string]*LRUNode
	head              *LRUNode // Most recently used
	tail              *LRUNode // Least recently used
	mu                sync.RWMutex
	defaultExpiration time.Duration
	cleanupInterval   time.Duration
	stopCleanup       chan bool
}

// LRUNode represents a node in the LRU cache
type LRUNode struct {
	key        string
	value      interface{}
	expiration int64
	prev       *LRUNode
	next       *LRUNode
}

// NewLRUCache creates a new LRU cache with the given capacity
func NewLRUCache(capacity int, defaultExpiration, cleanupInterval time.Duration) *LRUCache {
	cache := &LRUCache{
		capacity:          capacity,
		items:             make(map[string]*LRUNode, capacity),
		defaultExpiration: defaultExpiration,
		cleanupInterval:   cleanupInterval,
		stopCleanup:       make(chan bool),
	}
	
	// Start the cleanup goroutine
	if cleanupInterval > 0 {
		go cache.startCleanupTimer()
	}
	
	return cache
}

// startCleanupTimer starts the cleanup timer
func (c *LRUCache) startCleanupTimer() {
	ticker := time.NewTicker(c.cleanupInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			c.DeleteExpired()
		case <-c.stopCleanup:
			return
		}
	}
}

// Get gets an item from the cache and moves it to the front of the LRU list
func (c *LRUCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	node, found := c.items[key]
	c.mu.RUnlock()
	
	if !found {
		return nil, false
	}
	
	// Check if the item has expired
	if node.expiration > 0 && time.Now().UnixNano() > node.expiration {
		c.mu.Lock()
		c.removeNode(node)
		delete(c.items, key)
		c.mu.Unlock()
		return nil, false
	}
	
	// Move the node to the front of the list (most recently used)
	c.mu.Lock()
	c.moveToFront(node)
	c.mu.Unlock()
	
	return node.value, true
}

// Set adds an item to the cache with the default expiration
func (c *LRUCache) Set(key string, value interface{}) {
	c.SetWithExpiration(key, value, c.defaultExpiration)
}

// SetWithExpiration adds an item to the cache with a specific expiration
func (c *LRUCache) SetWithExpiration(key string, value interface{}, d time.Duration) {
	var expiration int64
	
	if d == 0 {
		// 0 means use default expiration
		d = c.defaultExpiration
	}
	
	if d > 0 {
		expiration = time.Now().Add(d).UnixNano()
	}
	
	c.mu.Lock()
	defer c.mu.Unlock()
	
	// Check if the key already exists
	if node, found := c.items[key]; found {
		// Update the value and expiration
		node.value = value
		node.expiration = expiration
		// Move the node to the front of the list
		c.moveToFront(node)
		return
	}
	
	// Create a new node
	node := &LRUNode{
		key:        key,
		value:      value,
		expiration: expiration,
	}
	
	// Add the node to the cache
	c.items[key] = node
	
	// Add the node to the front of the list
	if c.head == nil {
		// First node
		c.head = node
		c.tail = node
	} else {
		// Add to the front
		node.next = c.head
		c.head.prev = node
		c.head = node
	}
	
	// If the cache is over capacity, remove the least recently used item
	if len(c.items) > c.capacity {
		// Remove the tail node
		lru := c.tail
		c.removeNode(lru)
		delete(c.items, lru.key)
	}
}

// moveToFront moves a node to the front of the list
func (c *LRUCache) moveToFront(node *LRUNode) {
	if node == c.head {
		// Already at the front
		return
	}
	
	// Remove the node from its current position
	if node.prev != nil {
		node.prev.next = node.next
	}
	if node.next != nil {
		node.next.prev = node.prev
	}
	if node == c.tail {
		c.tail = node.prev
	}
	
	// Add the node to the front
	node.next = c.head
	node.prev = nil
	c.head.prev = node
	c.head = node
}

// removeNode removes a node from the linked list
func (c *LRUCache) removeNode(node *LRUNode) {
	if node.prev != nil {
		node.prev.next = node.next
	} else {
		// Node is the head
		c.head = node.next
	}
	
	if node.next != nil {
		node.next.prev = node.prev
	} else {
		// Node is the tail
		c.tail = node.prev
	}
}

// Delete deletes an item from the cache
func (c *LRUCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	node, found := c.items[key]
	if !found {
		return
	}
	
	c.removeNode(node)
	delete(c.items, key)
}

// DeleteExpired deletes all expired items from the cache
func (c *LRUCache) DeleteExpired() {
	now := time.Now().UnixNano()
	
	c.mu.Lock()
	defer c.mu.Unlock()
	
	for key, node := range c.items {
		if node.expiration > 0 && now > node.expiration {
			c.removeNode(node)
			delete(c.items, key)
		}
	}
}

// Flush deletes all items from the cache
func (c *LRUCache) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.items = make(map[string]*LRUNode, c.capacity)
	c.head = nil
	c.tail = nil
}

// Count returns the number of items in the cache
func (c *LRUCache) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	return len(c.items)
}

// Shutdown stops the cleanup goroutine
func (c *LRUCache) Shutdown() {
	if c.cleanupInterval > 0 {
		c.stopCleanup <- true
	}
}

// ConcurrentLRUCache implements a sharded LRU cache for better concurrency
type ConcurrentLRUCache struct {
	shards    []*LRUCache
	numShards int
}

// NewConcurrentLRUCache creates a new concurrent LRU cache with the given capacity
func NewConcurrentLRUCache(totalCapacity int, numShards int, defaultExpiration, cleanupInterval time.Duration) *ConcurrentLRUCache {
	if numShards <= 0 {
		numShards = 16 // Default number of shards
	}
	
	// Calculate capacity per shard
	shardCapacity := totalCapacity / numShards
	if shardCapacity < 1 {
		shardCapacity = 1
	}
	
	cache := &ConcurrentLRUCache{
		shards:    make([]*LRUCache, numShards),
		numShards: numShards,
	}
	
	// Create the shards
	for i := 0; i < numShards; i++ {
		cache.shards[i] = NewLRUCache(shardCapacity, defaultExpiration, cleanupInterval)
	}
	
	return cache
}

// getShard returns the shard for a given key
func (c *ConcurrentLRUCache) getShard(key string) *LRUCache {
	// Simple hash function to distribute keys
	hash := 0
	for i := 0; i < len(key); i++ {
		hash = 31*hash + int(key[i])
	}
	if hash < 0 {
		hash = -hash
	}
	return c.shards[hash%c.numShards]
}

// Get gets an item from the cache
func (c *ConcurrentLRUCache) Get(key string) (interface{}, bool) {
	return c.getShard(key).Get(key)
}

// Set adds an item to the cache with the default expiration
func (c *ConcurrentLRUCache) Set(key string, value interface{}) {
	c.getShard(key).Set(key, value)
}

// SetWithExpiration adds an item to the cache with a specific expiration
func (c *ConcurrentLRUCache) SetWithExpiration(key string, value interface{}, d time.Duration) {
	c.getShard(key).SetWithExpiration(key, value, d)
}

// Delete deletes an item from the cache
func (c *ConcurrentLRUCache) Delete(key string) {
	c.getShard(key).Delete(key)
}

// DeleteExpired deletes all expired items from the cache
func (c *ConcurrentLRUCache) DeleteExpired() {
	for i := 0; i < c.numShards; i++ {
		c.shards[i].DeleteExpired()
	}
}

// Flush deletes all items from the cache
func (c *ConcurrentLRUCache) Flush() {
	for i := 0; i < c.numShards; i++ {
		c.shards[i].Flush()
	}
}

// Count returns the number of items in the cache
func (c *ConcurrentLRUCache) Count() int {
	count := 0
	for i := 0; i < c.numShards; i++ {
		count += c.shards[i].Count()
	}
	return count
}

// Shutdown stops all cleanup goroutines
func (c *ConcurrentLRUCache) Shutdown() {
	for i := 0; i < c.numShards; i++ {
		c.shards[i].Shutdown()
	}
}
