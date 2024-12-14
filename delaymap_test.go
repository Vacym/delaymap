package delaymap_test

import (
	"testing"
	"time"

	"github.com/Vacym/delaymap"
)

func TestDelayMap_Set(t *testing.T) {
	dm := delaymap.New[string, int](time.Second)

	dm.Set("key1", 10)
	val, exists := dm.Get("key1")
	if !exists || val != 10 {
		t.Errorf("expected value 10, got %v", val)
	}
}

func TestDelayMap_SetAfterTimeout(t *testing.T) {
	dm := delaymap.New[string, int](time.Millisecond * 100)

	start := time.Now()
	go func() {
		time.Sleep(time.Millisecond * 500)
		dm.Set("key1", 10)
	}()

	val, exists := dm.Get("key1")
	elapsed := time.Since(start)

	if exists || val != 0 {
		t.Errorf("expected value 0, got %v", val)
	}

	if elapsed > time.Millisecond*200 {
		t.Errorf("Get took too long, elapsed time: %v", elapsed)
	}
}

func TestDelayMap_Get(t *testing.T) {
	dm := delaymap.New[string, int](time.Second)

	// Test getting a value that exists
	dm.Set("key1", 10)
	val, exists := dm.Get("key1")
	if !exists || val != 10 {
		t.Errorf("expected value 10, got %v", val)
	}

	// Test getting a value that does not exist
	val, exists = dm.Get("key2")
	if exists || val != 0 {
		t.Errorf("expected value 0, got %v", val)
	}
}

func TestDelayMap_GetEmpty(t *testing.T) {
	dm := delaymap.New[string, int](time.Second)

	// Test getting a value that exists
	val, exists := dm.Get("key1")
	if exists || val != 0 {
		t.Errorf("expected value 0, got %v", val)
	}
}

func TestDelayMap_GetWithTimeout(t *testing.T) {
	dm := delaymap.New[string, int](time.Millisecond * 100)

	// Test getting a value with timeout
	go func() {
		time.Sleep(time.Millisecond * 50)
		dm.Set("key1", 10)
	}()

	val, exists := dm.Get("key1")
	if !exists || val != 10 {
		t.Errorf("expected value 10, got %v", val)
	}

	// Test timeout
	val, exists = dm.Get("key2")
	if exists || val != 0 {
		t.Errorf("expected value 0, got %v", val)
	}
}

func TestDelayMap_GetWithTimeoutMultipleKeys(t *testing.T) {
	dm := delaymap.New[string, int](time.Millisecond * 100)

	// Test getting a value with timeout
	go func() {
		time.Sleep(time.Millisecond * 50)
		dm.Set("key1", 10)
	}()

	go func() {
		time.Sleep(time.Millisecond * 30)
		dm.Set("key2", 20)
	}()

	val, exists := dm.Get("key1")
	if !exists || val != 10 {
		t.Errorf("expected value 10, got %v", val)
	}

	// Test timeout
	go func() {
		time.Sleep(time.Millisecond * 200)
		dm.Set("key2", 20)
	}()

	val, exists = dm.Get("key2")
	if !exists || val != 20 {
		t.Errorf("expected value 20, got %v", val)
	}
}

func TestDelayMap_Delete(t *testing.T) {
	dm := delaymap.New[string, int](time.Second)

	dm.Set("key1", 10)
	dm.Delete("key1")
	val, exists := dm.Get("key1")
	if exists || val != 0 {
		t.Errorf("expected value 0, got %v", val)
	}
}
