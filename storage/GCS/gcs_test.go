package gcs

import (
	"os"
	"testing"

	"github.com/QingsiLiu/baseComponents/storage"
)

// TestNewGCSClient 测试 GCS 客户端创建
func TestNewGCSClient(t *testing.T) {
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		t.Skip("Skipping test: GOOGLE_CLOUD_PROJECT environment variable not set")
	}

	client, err := NewGCSClient(projectID, "")
	if err != nil {
		t.Fatalf("Failed to create GCS client: %v", err)
	}
	defer client.Close()

	if client.projectID != projectID {
		t.Errorf("Expected project ID %s, got %s", projectID, client.projectID)
	}
}

// TestGCSClientInterface 测试 GCS 客户端是否实现了 StorageService 接口
func TestGCSClientInterface(t *testing.T) {
	var _ storage.StorageService = (*GCSClient)(nil)
}