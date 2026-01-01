package server

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// MockS3Client implements S3ClientAPI
type MockS3Client struct {
	Objects map[string][]byte
}

func (m *MockS3Client) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	if m.Objects == nil {
		m.Objects = make(map[string][]byte)
	}
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(params.Body)
	m.Objects[*params.Key] = buf.Bytes()
	return &s3.PutObjectOutput{}, nil
}

func (m *MockS3Client) GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	if content, ok := m.Objects[*params.Key]; ok {
		return &s3.GetObjectOutput{
			Body: io.NopCloser(bytes.NewReader(content)),
		}, nil
	}
	// Simulate error
	// In real SDK it returns error types, but for now just returning error is enough to test Get logic
	// But Get logic checks err != nil.
	// We need to return an error if not found?
	// The current implementation returns err if GetObject fails.
	// So we should simulate failure if key missing.
	// But wait, the SDK returns NoSuchKey error.
	// For now, let's just return a generic error if not found.
	return nil, io.EOF // Just some error
}

func (m *MockS3Client) DeleteObject(_ context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	delete(m.Objects, *params.Key)
	return &s3.DeleteObjectOutput{}, nil
}

func TestS3BlobStore(t *testing.T) {
	mockClient := &MockS3Client{Objects: make(map[string][]byte)}
	store := &S3BlobStore{
		Client: mockClient,
		Bucket: "test-bucket",
	}

	// Test NewS3BlobStore (will fail in test due to real AWS config call, but we can test the struct directly)
	// Or we just skip the real constructor test and test the methods.

	id := "file1"
	content := []byte("content")

	// Test Save
	err := store.Save(id, content)
	if err != nil {
		t.Errorf("Save failed: %v", err)
	}
	if string(mockClient.Objects[id]) != string(content) {
		t.Error("Content not saved to mock")
	}

	// Test Get
	got, err := store.Get(id)
	if err != nil {
		t.Errorf("Get failed: %v", err)
	}
	if string(got) != string(content) {
		t.Errorf("Get mismatch")
	}

	// Test Delete
	if err := store.Delete(id); err != nil {
		t.Errorf("Delete failed: %v", err)
	}
	if _, ok := mockClient.Objects[id]; ok {
		t.Error("Object not deleted from mock")
	}
}
