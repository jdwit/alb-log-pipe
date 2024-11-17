package types

import "time"

type LogEntry struct {
	Data      map[string]string // Map of field name to value, this will be converted to JSON
	Timestamp time.Time
}
