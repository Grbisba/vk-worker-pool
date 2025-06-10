package pool

import (
	"time"
)

type Config struct {
	MaxReadAttempts  uint
	MaxWriteAttempts uint
	MaxWorkers       uint
	MaxInactivity    time.Duration
}
