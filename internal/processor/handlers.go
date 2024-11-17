package processor

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/jdwit/alb-log-pipe/internal/types"
	"log"
	"strings"
	"sync"
)

// concurrency is the max number of concurrent log processing operations
const concurrency = 10

func (lp *LogProcessor) processS3Objects(s3Objects []types.S3ObjectInfo) error {
	errs := make(chan error, len(s3Objects)) // buffered channel for errors
	var wg sync.WaitGroup
	concurrent := make(chan int, concurrency) // buffered channel for concurrency

	for _, s3obj := range s3Objects {
		wg.Add(1)
		concurrent <- 1
		go func(s3obj types.S3ObjectInfo) {
			defer func() {
				log.Printf("completed processing s3://%s/%s", s3obj.Bucket, s3obj.Key)
				wg.Done()
				<-concurrent
			}()
			err := lp.ProcessLogs(s3obj)
			if err != nil {
				errs <- fmt.Errorf("error processing logs for s3://%s/%s: %w", s3obj.Bucket, s3obj.Key, err)
			}
		}(s3obj)
	}

	go func() {
		wg.Wait()
		close(errs)
	}()

	var errorList []error
	for err := range errs {
		if err != nil {
			errorList = append(errorList, err)
		}
	}

	if len(errorList) > 0 {
		return fmt.Errorf("encountered errors: %v", errorList)
	}

	return nil
}

func (lp *LogProcessor) HandleLambdaEvent(event types.S3ObjectCreatedEvent) error {
	var s3Objects []types.S3ObjectInfo
	for _, record := range event.Records {
		s3Objects = append(s3Objects, types.S3ObjectInfo{
			Bucket: record.S3.Bucket.Name,
			Key:    record.S3.Object.Key,
		})
	}
	return lp.processS3Objects(s3Objects)
}

func (lp *LogProcessor) HandleS3URL(url string) error {
	bucket, prefix, err := parseS3Url(url)
	if err != nil {
		return fmt.Errorf("failed to parse S3 URL: %v", err)
	}

	var s3Objects []types.S3ObjectInfo
	var continuationToken *string
	for {
		resp, err := lp.s3Client.ListObjectsV2(&s3.ListObjectsV2Input{
			Bucket:            aws.String(bucket),
			Prefix:            aws.String(prefix),
			ContinuationToken: continuationToken,
		})
		if err != nil {
			return fmt.Errorf("failed to list objects: %v", err)
		}

		for _, item := range resp.Contents {
			s3Objects = append(s3Objects, types.S3ObjectInfo{
				Bucket: bucket,
				Key:    *item.Key,
			})
		}

		if resp.IsTruncated == nil || !*resp.IsTruncated {
			break
		}
		continuationToken = resp.NextContinuationToken
	}

	return lp.processS3Objects(s3Objects)
}

func parseS3Url(url string) (bucket string, prefix string, err error) {
	if !strings.HasPrefix(url, "s3://") {
		return "", "", fmt.Errorf("invalid S3 URL, missing 's3://' prefix")
	}
	trimmedS3URL := strings.TrimPrefix(url, "s3://")
	splitPos := strings.Index(trimmedS3URL, "/")
	if splitPos == -1 {
		return "", "", fmt.Errorf("invalid S3 URL, no '/' found after bucket name")
	}
	bucket = trimmedS3URL[:splitPos]
	prefix = trimmedS3URL[splitPos+1:]
	return bucket, prefix, nil
}
