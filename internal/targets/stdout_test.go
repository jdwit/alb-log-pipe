package targets

import (
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jdwit/alb-log-pipe/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestStdoutTarget_SendLogs(t *testing.T) {
	entryChan := make(chan types.LogEntry, 2)

	entryChan <- types.LogEntry{
		Timestamp: time.Date(2024, time.November, 17, 12, 0, 0, 0, time.UTC),
		Data:      map[string]string{"message": "test log 1"},
	}
	entryChan <- types.LogEntry{
		Timestamp: time.Date(2024, time.November, 17, 13, 0, 0, 0, time.UTC),
		Data:      map[string]string{"message": "test log 2"},
	}
	close(entryChan)

	r, w, _ := os.Pipe()
	originalStdout := os.Stdout
	os.Stdout = w
	defer func() {
		os.Stdout = originalStdout
		r.Close()
		w.Close()
	}()

	target := NewStdoutTarget()
	target.SendLogs(entryChan)
	w.Close()

	output, _ := io.ReadAll(r)
	actualOutput := strings.TrimSpace(string(output))

	expectedOutput := strings.Join([]string{
		`[2024-11-17T12:00:00Z] Log Entry: {"message":"test log 1"}`,
		`[2024-11-17T13:00:00Z] Log Entry: {"message":"test log 2"}`,
	}, "\n")

	assert.Equal(t, expectedOutput, actualOutput)
}
