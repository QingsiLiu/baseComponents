package modelslab

import (
	"os"
	"testing"
)

func requireModelsLabAPIKey(t *testing.T) {
	t.Helper()
	if os.Getenv(APIKeyEnvVar) == "" {
		t.Skip("Skipping integration test: MODELSLAB_API_KEY environment variable not set")
	}
}
