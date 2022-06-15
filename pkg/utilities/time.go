package utilities

import "sync"
import "time"

var lastTimeValue int64 = 0
var lastTimeLock sync.Mutex = sync.Mutex{}

func GetUniqueTime() int64 {
	var now int64

	lastTimeLock.Lock()
	defer lastTimeLock.Unlock()

	now = time.Now().UTC().UnixMicro()
	if now <= lastTimeValue {
		now = lastTimeValue + 1
	}
	lastTimeValue = now

	return now
}
