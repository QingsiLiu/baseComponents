package text2image

import "testing"

func TestServiceTypeMappings(t *testing.T) {
	if !IsValidSource(SourceKieGPTImage2Text2Image) {
		t.Fatalf("expected source %s to be valid", SourceKieGPTImage2Text2Image)
	}

	if got := GetServiceType(SourceKieGPTImage2Text2Image); got != ServiceTypeKieGPTImage2Text2Image {
		t.Fatalf("expected service type %s, got %s", ServiceTypeKieGPTImage2Text2Image, got)
	}

	if got := GetServiceSource(ServiceTypeKieGPTImage2Text2Image); got != SourceKieGPTImage2Text2Image {
		t.Fatalf("expected source %s, got %s", SourceKieGPTImage2Text2Image, got)
	}
}
