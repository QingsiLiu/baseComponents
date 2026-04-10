package kie

import "testing"

func requireKIEAPIKey(t *testing.T) {
	t.Helper()
	if GetAPIKey() == "" {
		t.Skip("Skipping integration test: KIE_API_KEY environment variable not set")
	}
}
