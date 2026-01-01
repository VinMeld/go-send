package server

import (
	"bytes"
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3ClientAPI defines the interface for S3 operations we use.
type S3ClientAPI interface {
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
}

// S3BlobStore implements BlobStore using AWS S3.
type S3BlobStore struct {
	Client S3ClientAPI
	Bucket string
}

func NewS3BlobStore(ctx context.Context, bucket string) (*S3BlobStore, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg)
	return &S3BlobStore{
		Client: client,
		Bucket: bucket,
	}, nil
}

func (s *S3BlobStore) Save(id string, content []byte) error {
	_, err := s.Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(id),
		Body:   bytes.NewReader(content),
	})
	return err
}

func (s *S3BlobStore) Get(id string) ([]byte, error) {
	resp, err := s.Client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(id),
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func (s *S3BlobStore) Delete(id string) error {
	_, err := s.Client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(id),
	})
	return err
}
