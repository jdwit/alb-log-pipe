package processor

import (
	"compress/gzip"
	"encoding/csv"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/jdwit/alb-log-pipe/internal/targets"
	"github.com/jdwit/alb-log-pipe/internal/types"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

type S3Api interface {
	GetObject(input *s3.GetObjectInput) (*s3.GetObjectOutput, error)
	ListObjectsV2(input *s3.ListObjectsV2Input) (*s3.ListObjectsV2Output, error)
}

type LogProcessor struct {
	s3Client S3Api
	config   Config
}

type Config struct {
	Targets []targets.Target
	Fields  Fields
}

const (
	// maxBatchCount The maximum number of events in a PutLogEvents request to CloudWatch is 10_000
	maxBatchCount = 10_000
)

func NewLogProcessor(sess *session.Session) (*LogProcessor, error) {
	s3Client := s3.New(sess)

	f, err := NewFields(os.Getenv("FIELDS"))
	if err != nil {
		return nil, err
	}

	t, err := targets.GetTargets(os.Getenv("TARGETS"), sess)
	if err != nil {
		return nil, err
	}

	cfg := Config{
		Targets: t,
		Fields:  f,
	}

	return &LogProcessor{
		s3Client: s3Client,
		config:   cfg,
	}, nil
}

func (lp *LogProcessor) ProcessLogs(s3Object types.S3ObjectInfo) error {
	log.Printf("processing logs from s3://%s/%s", s3Object.Bucket, s3Object.Key)

	obj, err := lp.s3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s3Object.Bucket),
		Key:    aws.String(s3Object.Key),
	})

	if err != nil {
		return fmt.Errorf("failed to get object: %v", err)
	}

	defer obj.Body.Close()

	reader, writer := io.Pipe()

	// Decompress the gzip file in a goroutine
	go func() {
		gzipReader, err := gzip.NewReader(obj.Body)
		if err != nil {
			writer.CloseWithError(err)

			return
		}
		defer gzipReader.Close()
		// Copy decompressed data to writer
		if _, err := io.Copy(writer, gzipReader); err != nil {
			writer.CloseWithError(err)

			return
		}
		writer.Close()
	}()

	// Set channel buffer size to 1.25 times the max batch count to avoid blocking
	entryChan := make(chan types.LogEntry, int(float64(maxBatchCount)*1.25))

	// Start each target in a separate goroutine
	var wg sync.WaitGroup
	for _, target := range lp.config.Targets {
		wg.Add(1)
		go func(t targets.Target) {
			defer wg.Done()
			t.SendLogs(entryChan)
		}(target)
	}

	// Process records and send to entryChan
	if err := processRecords(reader, entryChan, lp.config.Fields); err != nil {
		log.Printf("error processing records: %v", err)
	}

	close(entryChan)
	wg.Wait()

	return nil
}

func processRecords(reader io.Reader, entryChan chan types.LogEntry, fields Fields) error {
	csvReader := csv.NewReader(reader)
	csvReader.Comma = ' '
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading a record: %v", err)
		}
		entry, err := recordToLogEntry(record, fields)
		if err != nil {
			return err
		}
		entryChan <- entry
	}

	return nil
}

func recordToLogEntry(record []string, fields Fields) (types.LogEntry, error) {
	// Check if the record has the expected number of fields
	if len(record) != len(fieldNames) {
		return types.LogEntry{}, fmt.Errorf("invalid log format: expected %d fields, got %d", len(fieldNames), len(record))
	}
	timestamp, err := time.Parse(time.RFC3339, record[1]) // Timestamp should be at index 1
	if err != nil {
		return types.LogEntry{}, fmt.Errorf("error parsing timestamp: %v", err)
	}
	entryMap := make(map[string]string)
	for i, value := range record {
		// Only include the fields that we want
		if fields.IncludeField(i) {
			fieldName, _ := fields.GetFieldNameByIndex(i)
			entryMap[fieldName] = value
		}
	}

	return types.LogEntry{
		Data:      entryMap,
		Timestamp: timestamp,
	}, nil
}
