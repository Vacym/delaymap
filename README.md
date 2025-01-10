# DelayMap

A thread-safe generic map implementation in Go that supports delayed value retrieval with timeouts.

## Description

DelayMap is a concurrent-safe map structure that allows setting and getting values with a built-in waiting mechanism. When attempting to get a value that doesn't exist, DelayMap will wait for the specified timeout duration before returning, making it useful for scenarios where values might be set asynchronously.

## Installation

```sh
go get github.com/Vacym/delaymap
```

## Usage

Here's a simple example of how to use DelayMap:

```go
package main

import (
    "time"
    "fmt"
    "github.com/Vacym/delaymap"
)

func main() {
    // Create a new DelayMap with 1 second timeout
    dm := delaymap.New[string, int](time.Second)

    // Set value asynchronously
    go func() {
        time.Sleep(500 * time.Millisecond)
        dm.Set("key", 42)
    }()

    // Get will wait up to 1 second for the value
    if value, exists := dm.Get("key"); exists {
        fmt.Printf("Got value: %d\n", value)
    } else {
        fmt.Println("Value not found within timeout")
    }
}
```

## Features

- Generic type support for both keys and values
- Thread-safe operations
- Configurable timeout for value retrieval
- Clean API for setting, getting, and deleting values
- Proper cleanup with Close() method

## API

### `New[K comparable, V any](waitTimeout time.Duration) *DelayMap[K, V]`
Creates a new DelayMap with the specified timeout duration.

### `Set(key K, value V)`
Sets a value for the given key, notifying any goroutines waiting for this key.

### `Get(key K) (V, bool)`
Retrieves a value for the given key. If the value doesn't exist, waits until it is set or until the timeout expires.

### `Delete(key K)`
Removes the specified key and its value from the map.

### `Close()`
Cleans up resources and returns all waiting `Get` calls.

## Thread Safety

All operations on DelayMap are thread-safe, protected by mutual exclusion.

## Requirements

- Go 1.18 or later (for generics support)

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
