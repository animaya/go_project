package cache

import (
	"sync"
	"testing"
	"time"
)

func TestCache(t *testing.T) {
	// Create a new cache with a default expiration of 100ms and cleanup every 50ms
	cache := NewCache(100*time.Millisecond, 50*time.Millisecond)
	defer cache.Shutdown()
	
	// Test Set and Get
	cache.Set("key1", "value1")
	
	// Test that the item is in the cache
	if value, found := cache.Get("key1"); !found || value != "value1" {
		t.Errorf("Expected 'value1' for 'key1', got %v (found: %v)", value, found)
	}
	
	// Test that a non-existent key is not in the cache
	if _, found := cache.Get("key2"); found {
		t.Error("Expected 'key2' to not be found")
	}
	
	// Test expiration
	cache.SetWithExpiration("key3", "value3", 50*time.Millisecond)
	
	// The item should be available immediately
	if value, found := cache.Get("key3"); !found || value != "value3" {
		t.Errorf("Expected 'value3' for 'key3', got %v (found: %v)", value, found)
	}
	
	// Wait for the item to expire
	time.Sleep(60 * time.Millisecond)
	
	// The item should be expired
	if _, found := cache.Get("key3"); found {
		t.Error("Expected 'key3' to be expired")
	}
	
	// Test Delete
	cache.Set("key4", "value4")
	cache.Delete("key4")
	
	// The item should be deleted
	if _, found := cache.Get("key4"); found {
		t.Error("Expected 'key4' to be deleted")
	}
	
	// Test Flush
	cache.Set("key5", "value5")
	cache.Flush()
	
	// The cache should be empty
	if cache.Count() != 0 {
		t.Errorf("Expected cache to be empty, got %d items", cache.Count())
	}
	
	// Test automatic cleanup
	cache.SetWithExpiration("key6", "value6", 30*time.Millisecond)
	
	// Wait for the cleanup interval
	time.Sleep(100 * time.Millisecond)
	
	// The item should be automatically cleaned up
	if _, found := cache.Get("key6"); found {
		t.Error("Expected 'key6' to be automatically deleted by cleanup")
	}
}

func TestLRUCache(t *testing.T) {
	// Create a new LRU cache with a capacity of 3
	cache := NewLRUCache(3, 100*time.Millisecond, 50*time.Millisecond)
	defer cache.Shutdown()
	
	// Add items to the cache
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")
	
	// Check that all items are in the cache
	for i := 1; i <= 3; i++ {
		key := "key" + string(rune('0'+i))
		value := "value" + string(rune('0'+i))
		if v, found := cache.Get(key); !found || v != value {
			t.Errorf("Expected '%s' for '%s', got %v (found: %v)", value, key, v, found)
		}
	}
	
	// Add one more item, which should evict the least recently used item (key1)
	cache.Set("key4", "value4")
	
	// key1 should be evicted
	if _, found := cache.Get("key1"); found {
		t.Error("Expected 'key1' to be evicted")
	}
	
	// key2, key3, and key4 should still be in the cache
	for i := 2; i <= 4; i++ {
		key := "key" + string(rune('0'+i))
		value := "value" + string(rune('0'+i))
		if v, found := cache.Get(key); !found || v != value {
			t.Errorf("Expected '%s' for '%s', got %v (found: %v)", value, key, v, found)
		}
	}
	
	// Access key2, making key3 the least recently used
	cache.Get("key2")
	
	// Add one more item, which should evict key3
	cache.Set("key5", "value5")
	
	// key3 should be evicted
	if _, found := cache.Get("key3"); found {
		t.Error("Expected 'key3' to be evicted")
	}
	
	// key2, key4, and key5 should still be in the cache
	expected := map[string]string{
		"key2": "value2",
		"key4": "value4",
		"key5": "value5",
	}
	
	for key, expectedValue := range expected {
		if value, found := cache.Get(key); !found || value != expectedValue {
			t.Errorf("Expected '%s' for '%s', got %v (found: %v)", expectedValue, key, value, found)
		}
	}
	
	// Test expiration
	cache.SetWithExpiration("key6", "value6", 50*time.Millisecond)
	
	// The item should be available immediately
	if value, found := cache.Get("key6"); !found || value != "value6" {
		t.Errorf("Expected 'value6' for 'key6', got %v (found: %v)", value, found)
	}
	
	// Wait for the item to expire
	time.Sleep(60 * time.Millisecond)
	
	// The item should be expired
	if _, found := cache.Get("key6"); found {
		t.Error("Expected 'key6' to be expired")
	}
	
	// Test Delete
	cache.Set("key7", "value7")
	cache.Delete("key7")
	
	// The item should be deleted
	if _, found := cache.Get("key7"); found {
		t.Error("Expected 'key7' to be deleted")
	}
	
	// Test Flush
	cache.Flush()
	
	// The cache should be empty
	if cache.Count() != 0 {
		t.Errorf("Expected cache to be empty, got %d items", cache.Count())
	}
}

func TestConcurrentLRUCache(t *testing.T) {
	// Create a new concurrent LRU cache with 100 total capacity spread across 4 shards
	cache := NewConcurrentLRUCache(100, 4, 100*time.Millisecond, 50*time.Millisecond)
	defer cache.Shutdown()
	
	// Test basic operations
	cache.Set("key1", "value1")
	
	if value, found := cache.Get("key1"); !found || value != "value1" {
		t.Errorf("Expected 'value1' for 'key1', got %v (found: %v)", value, found)
	}
	
	// Test concurrency
	var wg sync.WaitGroup
	numGoroutines := 100
	numOperations := 100
	
	// Prepare keys and values
	keys := make([]string, numOperations)
	values := make([]string, numOperations)
	for i := 0; i < numOperations; i++ {
		keys[i] = "key" + string(rune('a'+i%26)) + string(rune('0'+i/26))
		values[i] = "value" + string(rune('a'+i%26)) + string(rune('0'+i/26))
	}
	
	// Launch goroutines to set values
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			// Each goroutine sets a different subset of keys
			start := (id * numOperations) / numGoroutines
			end := ((id + 1) * numOperations) / numGoroutines
			
			for j := start; j < end; j++ {
				cache.Set(keys[j], values[j])
			}
		}(i)
	}
	
	// Wait for all set operations to complete
	wg.Wait()
	
	// Launch goroutines to get values
	errors := make(chan string, numGoroutines)
	
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			// Each goroutine gets a different subset of keys
			start := (id * numOperations) / numGoroutines
			end := ((id + 1) * numOperations) / numGoroutines
			
			for j := start; j < end; j++ {
				value, found := cache.Get(keys[j])
				if !found {
					errors <- "Key not found: " + keys[j]
				} else if value != values[j] {
					errors <- "Incorrect value for key " + keys[j] + ": expected " + values[j] + ", got " + value.(string)
				}
			}
		}(i)
	}
	
	// Wait for all get operations to complete
	wg.Wait()
	close(errors)
	
	// Check for errors
	for err := range errors {
		t.Error(err)
	}
	
	// Test that the count is correct
	// Some keys might be evicted if they map to the same shard and exceed the shard capacity
	count := cache.Count()
	if count == 0 {
		t.Error("Expected cache to contain items, but it was empty")
	}
	
	// Test Flush
	cache.Flush()
	
	// The cache should be empty
	if cache.Count() != 0 {
		t.Errorf("Expected cache to be empty after flush, got %d items", cache.Count())
	}
}
