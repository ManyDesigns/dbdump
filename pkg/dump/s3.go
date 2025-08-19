package dump

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

// Define the S3 struct.
type S3Uploader struct {
	bucket string
	region string
}

// Create an uploader reading environment variables to authenticate to AWS services
func NewS3Uploader(localDump bool) (*S3Uploader, error) {
	if localDump == false {
		bucket := os.Getenv("AWS_BUCKET")
		if bucket == "" {
			return nil, fmt.Errorf("environment variable 'AWS_BUCKET' is not set")
		}

		region := os.Getenv("AWS_REGION")
		if region == "" {
			return nil, fmt.Errorf("environment variable 'AWS_REGION' is not set")
		}

		return &S3Uploader{
			bucket: bucket,
			region: region,
		}, nil
	} else {
		return &S3Uploader{}, nil
	}
}

func (s *S3Uploader) Upload(localPath string, remotePath string) (string, error) {
	s3Uri := fmt.Sprintf("s3://%s/%s", s.bucket, remotePath)
	// We'll use AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY from the environment
	cmd := exec.Command("aws", "s3", "cp", localPath, s3Uri, "--region", s.region)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to upload to S3: %s", stderr.String())
	}

	return s3Uri, nil
}

func (s *S3Uploader) Download(remotePath string, localPath string) (string, error) {
	s3Uri := fmt.Sprintf("s3://%s/%s", s.bucket, remotePath)
	cmd := exec.Command("aws", "s3", "cp", s3Uri, localPath, "--region", s.region)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to download '%s' from S3 Bucket: %s", s3Uri, stderr.String())
	}
	return localPath, nil
}
