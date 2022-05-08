package godisson

import (
	"time"
)

func currentTimeMillis() int64 {
	return time.Now().UnixNano() / 1e6
}
