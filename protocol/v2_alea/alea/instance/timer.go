package instance

import (
	"fmt"
	"sync"
	"time"
)

func makeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Microsecond)
}

type Timer struct{}

var (
	mu      sync.Mutex
	timeMap = make(map[string]int64)
)

func (t *Timer) startTime(key string) {
	mu.Lock()
	defer mu.Unlock()
	if timeMap == nil {
		timeMap = make(map[string]int64)
	}
	timeMap[key] = makeTimestamp()
}

func (t *Timer) endTime(key string) int64 {
	mu.Lock()
	defer mu.Unlock()
	if val, ok := timeMap[key]; ok {
		// delete(timeMap, key) // remove the key from the map
		return makeTimestamp() - val
	}
	return -1 // or any other error value
}

func main() {
	t := Timer{}
	t.startTime("foo")
	time.Sleep(1 * time.Second)
	fmt.Println(t.endTime("foo"))
}
