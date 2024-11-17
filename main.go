package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/jdwit/alb-log-pipe/internal/processor"
	"log"
	"os"
)

func createSession() (*session.Session, error) {
	endpoint := os.Getenv("AWS_ENDPOINT")
	if endpoint != "" {
		// localstack
		return session.NewSession(&aws.Config{
			Endpoint:         aws.String(endpoint),
			DisableSSL:       aws.Bool(true),
			S3ForcePathStyle: aws.Bool(true),
		})
	}

	return session.NewSession()
}

func main() {
	sess, err := createSession()
	if err != nil {
		log.Fatalln(err)
	}

	lp, err := processor.NewLogProcessor(sess)
	if err != nil {
		log.Fatalln(err)
	}

	if os.Getenv("AWS_LAMBDA_RUNTIME_API") != "" {
		log.Println("running in AWS Lambda environment")
		lambda.Start(lp.HandleLambdaEvent)
	} else {
		log.Println("running in cli mode")
		if len(os.Args) < 2 {
			log.Fatalln("s3 url is required as an argument")
		}
		err := lp.HandleS3URL(os.Args[1])
		if err != nil {
			log.Fatalln(err)
		}
	}
}
