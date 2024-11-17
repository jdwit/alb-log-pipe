package targets

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockCloudWatchLogsClient struct {
	mock.Mock
}

func (m *MockCloudWatchLogsClient) PutLogEvents(input *cloudwatchlogs.PutLogEventsInput) (*cloudwatchlogs.PutLogEventsOutput, error) {
	args := m.Called(input)
	return args.Get(0).(*cloudwatchlogs.PutLogEventsOutput), args.Error(1)
}

func (m *MockCloudWatchLogsClient) CreateLogGroup(input *cloudwatchlogs.CreateLogGroupInput) (*cloudwatchlogs.CreateLogGroupOutput, error) {
	args := m.Called(input)
	return args.Get(0).(*cloudwatchlogs.CreateLogGroupOutput), args.Error(1)
}

func (m *MockCloudWatchLogsClient) CreateLogStream(input *cloudwatchlogs.CreateLogStreamInput) (*cloudwatchlogs.CreateLogStreamOutput, error) {
	args := m.Called(input)
	return args.Get(0).(*cloudwatchlogs.CreateLogStreamOutput), args.Error(1)
}

func (m *MockCloudWatchLogsClient) DescribeLogGroups(input *cloudwatchlogs.DescribeLogGroupsInput) (*cloudwatchlogs.DescribeLogGroupsOutput, error) {
	args := m.Called(input)
	return args.Get(0).(*cloudwatchlogs.DescribeLogGroupsOutput), args.Error(1)
}

func (m *MockCloudWatchLogsClient) DescribeLogStreams(input *cloudwatchlogs.DescribeLogStreamsInput) (*cloudwatchlogs.DescribeLogStreamsOutput, error) {
	args := m.Called(input)
	return args.Get(0).(*cloudwatchlogs.DescribeLogStreamsOutput), args.Error(1)
}

func TestEnsureLogGroupAndLogStreamExists(t *testing.T) {

	logConfig := LogConfig{
		LogGroupName:  "test-log-group",
		LogStreamName: "test-log-stream",
	}

	t.Run("Log group and stream exist", func(t *testing.T) {
		mockClient := new(MockCloudWatchLogsClient)
		mockClient.On("DescribeLogGroups", mock.Anything).Return(&cloudwatchlogs.DescribeLogGroupsOutput{
			LogGroups: []*cloudwatchlogs.LogGroup{
				{LogGroupName: aws.String("test-log-group")},
			},
		}, nil)

		mockClient.On("DescribeLogStreams", mock.Anything).Return(&cloudwatchlogs.DescribeLogStreamsOutput{
			LogStreams: []*cloudwatchlogs.LogStream{
				{LogStreamName: aws.String("test-log-stream")},
			},
		}, nil)

		err := ensureLogGroupAndLogStreamExists(mockClient, logConfig)
		require.NoError(t, err)

		mockClient.AssertExpectations(t)
	})

	t.Run("Log group does not exist", func(t *testing.T) {
		mockClient := new(MockCloudWatchLogsClient)
		mockClient.On("DescribeLogGroups", mock.Anything).Return(&cloudwatchlogs.DescribeLogGroupsOutput{
			LogGroups: []*cloudwatchlogs.LogGroup{},
		}, nil)

		mockClient.On("CreateLogGroup", &cloudwatchlogs.CreateLogGroupInput{
			LogGroupName: aws.String("test-log-group"),
		}).Return(&cloudwatchlogs.CreateLogGroupOutput{}, nil)

		mockClient.On("DescribeLogStreams", &cloudwatchlogs.DescribeLogStreamsInput{
			LogGroupName: aws.String("test-log-group"),
		}).Return(&cloudwatchlogs.DescribeLogStreamsOutput{
			LogStreams: []*cloudwatchlogs.LogStream{
				{LogStreamName: aws.String("test-log-stream")},
			},
		}, nil)

		err := ensureLogGroupAndLogStreamExists(mockClient, logConfig)
		require.NoError(t, err)

		mockClient.AssertExpectations(t)
	})

	t.Run("Log stream does not exist", func(t *testing.T) {
		mockClient := new(MockCloudWatchLogsClient)
		mockClient.On("DescribeLogGroups", mock.Anything).Return(&cloudwatchlogs.DescribeLogGroupsOutput{
			LogGroups: []*cloudwatchlogs.LogGroup{
				{LogGroupName: aws.String("test-log-group")},
			},
		}, nil)

		mockClient.On("DescribeLogStreams", &cloudwatchlogs.DescribeLogStreamsInput{
			LogGroupName: aws.String("test-log-group"),
		}).Return(&cloudwatchlogs.DescribeLogStreamsOutput{
			LogStreams: []*cloudwatchlogs.LogStream{},
		}, nil)

		mockClient.On("CreateLogStream", &cloudwatchlogs.CreateLogStreamInput{
			LogGroupName:  aws.String("test-log-group"),
			LogStreamName: aws.String("test-log-stream"),
		}).Return(&cloudwatchlogs.CreateLogStreamOutput{}, nil)

		err := ensureLogGroupAndLogStreamExists(mockClient, logConfig)
		require.NoError(t, err)

		mockClient.AssertExpectations(t)
	})
}

func TestSendEventsToCloudWatch(t *testing.T) {
	logConfig := LogConfig{
		LogGroupName:  "test-log-group",
		LogStreamName: "test-log-stream",
	}

	t.Run("Send events successfully", func(t *testing.T) {
		mockClient := new(MockCloudWatchLogsClient)
		mockClient.On("PutLogEvents", mock.Anything).Return(&cloudwatchlogs.PutLogEventsOutput{}, nil)

		events := []*cloudwatchlogs.InputLogEvent{
			{
				Message:   aws.String("message1"),
				Timestamp: aws.Int64(1),
			},
			{
				Message:   aws.String("message2"),
				Timestamp: aws.Int64(2),
			},
		}

		cwTarget := &CloudWatchTarget{
			cwClient:  mockClient,
			logConfig: logConfig,
		}

		cwTarget.sendBatch(events)

		mockClient.AssertExpectations(t)
	})
}
