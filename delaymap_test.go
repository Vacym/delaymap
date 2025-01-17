package delaymap_test

import (
	"sync"
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

func TestDelayMap_GetWithTimeoutRepeat(t *testing.T) {
	dm := delaymap.New[string, int](time.Millisecond * 100)

	var wg sync.WaitGroup

	go func() {
		time.Sleep(time.Millisecond * 50)
		dm.Set("key1", 10)
	}()

	tests := 5
	wg.Add(tests)

	for i := 0; i < tests; i++ {
		go func() {
			defer wg.Done()
			val, exists := dm.Get("key1")
			if !exists || val != 10 {
				t.Errorf("expected value 10, got %v", val)
			}
		}()
	}

	wg.Wait()
}

func TestDelayMap_GetWithTimeoutPartiallyGot(t *testing.T) {
	dm := delaymap.New[string, int](time.Millisecond * 50)

	var wg sync.WaitGroup

	go func() {
		time.Sleep(time.Millisecond * 100) // Устанавливаем ключ через 100 мс
		dm.Set("key1", 10)
	}()

	tests := 10
	interval := 20 // Проверяем каждые 20 мс
	wg.Add(tests)

	for i := 0; i < tests; i++ {
		timeFromStart := interval * i // Вычисляем время перед запуском горутины
		go func(i, timeFromStart int) {
			defer wg.Done()
			val, exists := dm.Get("key1")

			if timeFromStart >= 50 {
				if !exists || val != 10 {
					t.Errorf("expected value 10, got %v", val)
				}
			} else {
				if exists || val != 0 {
					t.Errorf("expected no value, got %v", val)
				}
			}
		}(i, timeFromStart)

		time.Sleep(time.Millisecond * time.Duration(interval))
	}

	wg.Wait()
}

func TestDelayMap_MultipleInteractions(t *testing.T) {
	dm := delaymap.New[string, int](time.Millisecond * 100)

	// Set multiple keys
	go func() {
		time.Sleep(time.Millisecond * 50)
		dm.Set("key1", 10)
		dm.Set("key2", 20)
		dm.Set("key3", 30)
	}()

	// Get multiple keys
	val, exists := dm.Get("key1")
	if !exists || val != 10 {
		t.Errorf("expected value 10, got %v", val)
	}

	val, exists = dm.Get("key2")
	if !exists || val != 20 {
		t.Errorf("expected value 20, got %v", val)
	}

	val, exists = dm.Get("key3")
	if !exists || val != 30 {
		t.Errorf("expected value 30, got %v", val)
	}

	// Delete a key and check
	dm.Delete("key2")
	val, exists = dm.Get("key2")
	if exists || val != 0 {
		t.Errorf("expected value 0, got %v", val)
	}

	// Set a key after deletion
	dm.Set("key2", 40)
	val, exists = dm.Get("key2")
	if !exists || val != 40 {
		t.Errorf("expected value 40, got %v", val)
	}
}
