package delaymap

import (
	"sync"
	"time"
)

// DelayMap is a generic map that allows setting and retrieving values with a timeout mechanism.
// If a value for a key is not immediately available, it waits for the specified timeout duration before returning.
type DelayMap[K comparable, V any] struct {
	data        map[K]V             // Internal map for storing key-value pairs.
	waitTimeout time.Duration       // Maximum time to wait for a key to be set.
	waitChans   map[K]chan struct{} // Channels to signal when a value becomes available for a key.
	mu          sync.Mutex          // Mutex to protect concurrent access.
}

// New creates a new DelayMap with the specified wait timeout.
func New[K comparable, V any](waitTimeout time.Duration) *DelayMap[K, V] {
	m := &DelayMap[K, V]{
		data:        make(map[K]V),
		waitTimeout: waitTimeout,
		waitChans:   make(map[K]chan struct{}),
	}
	return m
}

// Set assigns a value to the given key and signals any waiting goroutines.
func (dm *DelayMap[K, V]) Set(key K, value V) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	dm.data[key] = value
	if ch, exists := dm.waitChans[key]; exists {
		close(ch) // Signal waiting goroutines.
		delete(dm.waitChans, key)
	}
}

// Get retrieves the value for the given key. If the value is not available, it waits for the specified timeout.
// Returns the value and true if the key exists, or the zero value and false otherwise.
func (dm *DelayMap[K, V]) Get(key K) (V, bool) {
	dm.mu.Lock()
	if val, exists := dm.data[key]; exists {
		dm.mu.Unlock()
		return val, true
	}
	ch := make(chan struct{})
	dm.waitChans[key] = ch
	dm.mu.Unlock()

	select {
	case <-ch:
		dm.mu.Lock()
		val, exists := dm.data[key]
		dm.mu.Unlock()
		return val, exists
	case <-time.After(dm.waitTimeout):
		dm.mu.Lock()
		delete(dm.waitChans, key) // Clean up wait channel.
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

// Close clears the map and ends all pending expectations by closing all wait channels.
func (dm *DelayMap[K, V]) Close() {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	for key, ch := range dm.waitChans {
		close(ch) // Notify all waiting goroutines.
		delete(dm.waitChans, key)
	}
	dm.data = make(map[K]V)                  // Reset the map.
	dm.waitChans = make(map[K]chan struct{}) // Reset the wait channels.
}
