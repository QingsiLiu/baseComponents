package replicate

import "testing"

func requireReplicateToken(t *testing.T) {
	t.Helper()
	if GetAPIToken() == "" {
		t.Skip("Skipping integration test: REPLICATE_TOKEN environment variable not set")
	}
}
