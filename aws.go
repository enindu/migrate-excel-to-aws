package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func createAwsS3Uploader() *s3manager.Uploader {
	awsSession, exception := createAwsSession()
	handle(exception)

	return s3manager.NewUploader(awsSession)
}

func createAwsSession() (*session.Session, error) {
	return session.NewSession(&aws.Config{
		Region:      aws.String(AWS_REGION),
		Credentials: credentials.NewStaticCredentials(AWS_ACCESS_KEY, AWS_SECRET_KEY, ""),
	})
}
