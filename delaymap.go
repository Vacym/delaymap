package delaymap

import (
	"sync"
	"time"
)

// DelayMap is a generic map that allows setting and retrieving values with a timeout mechanism.
// If a value for a key is not immediately available, it waits for the specified timeout duration before returning.
type DelayMap[K comparable, V any] struct {
	data        map[K]V          // Internal map for storing key-value pairs.
	waitTimeout time.Duration    // Maximum time to wait for a key to be set.
	conds       map[K]*sync.Cond // Condition variables for waiting goroutines.
	mu          sync.Mutex       // Mutex to protect concurrent access.
}

// New creates a new DelayMap with the specified wait timeout.
func New[K comparable, V any](waitTimeout time.Duration) *DelayMap[K, V] {
	return &DelayMap[K, V]{
		data:        make(map[K]V),
		waitTimeout: waitTimeout,
		conds:       make(map[K]*sync.Cond),
	}
}

// Set assigns a value to the given key and signals any waiting goroutines.
func (dm *DelayMap[K, V]) Set(key K, value V) {
	dm.mu.Lock()
	dm.data[key] = value
	if cond, exists := dm.conds[key]; exists {
		cond.Broadcast() // Wake up all waiting goroutines.
		delete(dm.conds, key)
	}
	dm.mu.Unlock()
}

// Get retrieves the value for the given key. If the value is not available, it waits for the specified timeout.
// Returns the value and true if the key exists, or the zero value and false otherwise.
func (dm *DelayMap[K, V]) Get(key K) (V, bool) {
	dm.mu.Lock()
	if val, exists := dm.data[key]; exists {
		dm.mu.Unlock()
		return val, true
	}

	// Create a new sync.Cond if it does not exist.
	cond, exists := dm.conds[key]
	if !exists {
		cond = sync.NewCond(&dm.mu)
		dm.conds[key] = cond
	}
	dm.mu.Unlock()

	// Wait for the value to be set or timeout.
	timeout := time.After(dm.waitTimeout)
	done := make(chan struct{}) // Channel for synchronization.

	go func() {
		dm.mu.Lock()
		defer dm.mu.Unlock()
		cond.Wait() // Wait for a signal from Set.
		close(done)
	}()

	select {
	case <-done:
		dm.mu.Lock()
		val, exists := dm.data[key]
		dm.mu.Unlock()
		return val, exists
	case <-timeout:
		dm.mu.Lock()
		dm.mu.Unlock()
		var zero V
		return zero, false
	}
}

// Delete removes the given key from the map.
func (dm *DelayMap[K, V]) Delete(key K) {
	dm.mu.Lock()
	delete(dm.data, key)
	dm.mu.Unlock()
}

// Close clears the map and ends all pending expectations by broadcasting all wait conditions.
func (dm *DelayMap[K, V]) Close() {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	for key, cond := range dm.conds {
		cond.Broadcast() // Notify all waiting goroutines.
		delete(dm.conds, key)
	}
	dm.data = make(map[K]V)           // Reset the map.
	dm.conds = make(map[K]*sync.Cond) // Reset the condition variables.
}
