package targets

import (
	"encoding/json"
	"fmt"
	"github.com/jdwit/alb-log-pipe/internal/types"
	"time"
)

type StdoutTarget struct{}

func (c *StdoutTarget) SendLogs(entryChan <-chan types.LogEntry) {
	for entry := range entryChan {
		jsonData, err := json.Marshal(entry.Data)
		if err != nil {
			fmt.Printf("error marshaling log entry to JSON: %v\n", err)
			continue
		}
		fmt.Printf("[%s] Log Entry: %s\n", entry.Timestamp.Format(time.RFC3339), jsonData)
	}
}

func NewStdoutTarget() *StdoutTarget {
	return &StdoutTarget{}
}
