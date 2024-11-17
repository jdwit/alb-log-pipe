package targets

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/jdwit/alb-log-pipe/internal/types"
	"log"
	"os"
	"sort"
	"time"
)

const (
	// maxBatchSize The maximum batch size of a PutLogEvents request to CloudWatch is 1MB (1_048_576 bytes)
	maxBatchSize = 1_048_576
	// maxBatchCount The maximum number of events in a PutLogEvents request to CloudWatch is 10_000
	maxBatchCount = 10_000
)

type CloudWatchLogsAPI interface {
	PutLogEvents(*cloudwatchlogs.PutLogEventsInput) (*cloudwatchlogs.PutLogEventsOutput, error)
	CreateLogGroup(*cloudwatchlogs.CreateLogGroupInput) (*cloudwatchlogs.CreateLogGroupOutput, error)
	CreateLogStream(*cloudwatchlogs.CreateLogStreamInput) (*cloudwatchlogs.CreateLogStreamOutput, error)
	DescribeLogGroups(*cloudwatchlogs.DescribeLogGroupsInput) (*cloudwatchlogs.DescribeLogGroupsOutput, error)
	DescribeLogStreams(*cloudwatchlogs.DescribeLogStreamsInput) (*cloudwatchlogs.DescribeLogStreamsOutput, error)
}

type LogConfig struct {
	LogGroupName  string
	LogStreamName string
}

type CloudWatchTarget struct {
	cwClient  CloudWatchLogsAPI
	logConfig LogConfig
}

func (c *CloudWatchTarget) SendLogs(entryChan <-chan types.LogEntry) {
	var events []*cloudwatchlogs.InputLogEvent
	var currentBatchSize int

	ticker := time.NewTicker(5 * time.Second) // Send any remaining logs every 5 seconds
	defer ticker.Stop()

	for {
		select {
		case entry, ok := <-entryChan:
			if !ok {
				// Channel closed, send remaining events
				if len(events) > 0 {
					c.sendBatch(events)
				}
				return
			}

			jsonData, err := json.Marshal(entry.Data)
			if err != nil {
				fmt.Println("error marshaling log entry to JSON:", err)
				continue
			}
			event := &cloudwatchlogs.InputLogEvent{
				Message:   aws.String(string(jsonData)),
				Timestamp: aws.Int64(entry.Timestamp.UnixMilli()),
			}

			// Request size to CloudWatch is calculated as the sum of all event messages in UTF-8, plus 26 bytes for each log event
			// https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/cloudwatch_limits_cwl.html
			eventSize := len(jsonData) + 26

			// Check if adding this event exceeds batch size or count
			if len(events) > 0 && (currentBatchSize+eventSize > maxBatchSize || len(events) >= maxBatchCount) {
				c.sendBatch(events)
				events = nil
				currentBatchSize = 0
			}

			// Add event to batch
			events = append(events, event)
			currentBatchSize += eventSize

		case <-ticker.C:
			// Send remaining events every 5 seconds, this ensures logs are sent even if batch size is not reached
			if len(events) > 0 {
				c.sendBatch(events)
				events = nil
				currentBatchSize = 0
			}
		}
	}
}

func NewCloudWatchTarget(sess *session.Session) (Target, error) {
	logGroupName := os.Getenv("CLOUDWATCH_LOG_GROUP")
	if logGroupName == "" {
		return nil, fmt.Errorf("environment variable CLOUDWATCH_LOG_GROUP is required")
	}

	logStreamName := os.Getenv("CLOUDWATCH_LOG_STREAM")
	if logStreamName == "" {
		return nil, fmt.Errorf("environment variable CLOUDWATCH_LOG_STREAM is required")
	}

	logConfig := LogConfig{
		LogGroupName:  logGroupName,
		LogStreamName: logStreamName,
	}

	cwClient := cloudwatchlogs.New(sess)
	err := ensureLogGroupAndLogStreamExists(cwClient, logConfig)

	if err != nil {
		return nil, fmt.Errorf("error creating log group and stream: %v", err)
	}

	return &CloudWatchTarget{cwClient: cwClient, logConfig: logConfig}, nil
}

func ensureLogGroupAndLogStreamExists(client CloudWatchLogsAPI, logConfig LogConfig) error {
	err := ensureLogGroupExists(client, logConfig.LogGroupName)
	if err != nil {
		return err
	}
	err = ensureLogStreamExists(client, logConfig.LogGroupName, logConfig.LogStreamName)

	return err
}

func ensureLogGroupExists(client CloudWatchLogsAPI, name string) error {
	resp, err := client.DescribeLogGroups(&cloudwatchlogs.DescribeLogGroupsInput{})
	if err != nil {
		return err
	}
	for _, logGroup := range resp.LogGroups {
		if *logGroup.LogGroupName == name {
			return nil
		}
	}
	log.Printf("creating log group %s", name)
	_, err = client.CreateLogGroup(&cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: aws.String(name),
	})

	return err
}

func ensureLogStreamExists(client CloudWatchLogsAPI, logGroupName, logStreamName string) error {
	resp, err := client.DescribeLogStreams(&cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName: aws.String(logGroupName),
	})
	if err != nil {
		return err
	}
	for _, logStream := range resp.LogStreams {
		if *logStream.LogStreamName == logStreamName {
			return nil
		}
	}
	log.Printf("creating log stream %s in log group %s", logStreamName, logGroupName)
	_, err = client.CreateLogStream(&cloudwatchlogs.CreateLogStreamInput{
		LogGroupName:  aws.String(logGroupName),
		LogStreamName: aws.String(logStreamName),
	})

	return err
}

func (c *CloudWatchTarget) sendBatch(events []*cloudwatchlogs.InputLogEvent) {
	// Log events in a single PutLogEvents request must be in chronological order
	sort.Slice(events, func(i, j int) bool {
		return aws.Int64Value(events[i].Timestamp) < aws.Int64Value(events[j].Timestamp)
	})
	_, err := c.cwClient.PutLogEvents(&cloudwatchlogs.PutLogEventsInput{
		LogEvents:     events,
		LogGroupName:  aws.String(c.logConfig.LogGroupName),
		LogStreamName: aws.String(c.logConfig.LogStreamName),
	})

	if err != nil {
		fmt.Println("error sending events to CloudWatch:", err)
	}
}
