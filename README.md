# ALB Log Pipe

Process and deliver your AWS Application Load Balancer access logs anywhere!

You can configure Application Load Balancers
to [store access logs in an S3 bucket](https://docs.aws.amazon.com/elasticloadbalancing/latest/application/enable-access-logging.html).
However, handling large tarballs with raw log files is cumbersome. ALB Log Pipe can be installed
as a Lambda function to process the raw logs stored in S3 and forward them to various targets.

## Supported targets

- [x] CloudWatch Logs
- [x] stdout
- [ ] TODO Logstash

Do you need another target? Open an issue or submit a PR!

## Configuration

Configuration is done via environment variables.

### Common Configuration
- `FIELDS`: List of comma separated fields to extract from the log line. If not provided, all fields will be sent by
  default. For a list of all available fields
  see [ALB docs](https://docs.aws.amazon.com/elasticloadbalancing/latest/application/load-balancer-access-logs.html#access-log-entry-format)
- `TARGETS`: List of comma separated targets to send logs to. Supported targets are `cloudwatch`, `stdout`. At least one
  target must be provided.

### Target Configuration

| Target              | Identifier   | Variables                                                                                                                                     |
|---------------------|--------------|-----------------------------------------------------------------------------------------------------------------------------------------------|
| **CloudWatch Logs** | `cloudwatch` | `CLOUDWATCH_LOG_GROUP`: CloudWatch Log Group Name to send logs to. <br/> `CLOUDWATCH_LOG_STREAM`: CloudWatch Log Stream Name to send logs to. |
| **stdout**          | `stdout`     | No additional configuration required.                                                                                                         |

## CLI Usage

The program can be run from the command line to process log files stored in S3. A S3 URI is expected as the only
argument and this URI should point to a directory containing ALB log files.

For example, you want to process all log files stored for January 1st, 2024, and send them to CloudWatch. You are only
interested in the request URL and the response processing time. You can do this by running:

```
TARGETS=cloudwatch \
CLOUDWATCH_LOG_GROUP=my-log-group-name \
CLOUDWATCH_LOG_STREAM=my-log-stream-name \
FIELDS=request,response_processing_time \
./alb-log-pipe s3://<bucket>/AWSLogs/<account-id>/elasticloadbalancing/<region>/2024/01/01/
```

## Lambda Function

The tool is primarily intended to be used as a Lambda function. By configuring the function to be triggered by an
`s3:ObjectCreated` event, log files are processed and sent to the configured targets as soon as they are stored in S3. 
A SAM template is provided.