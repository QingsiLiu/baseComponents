package kling

import (
	"reflect"
	"testing"
	"unsafe"

	aivideokling "github.com/QingsiLiu/baseComponents/service/aivideo/kling"
	providerwellapi "github.com/QingsiLiu/baseComponents/service/thirdparty/wellapi"
)

func readStringField(t *testing.T, v reflect.Value) string {
	t.Helper()
	ptr := unsafe.Pointer(v.UnsafeAddr())
	return reflect.NewAt(v.Type(), ptr).Elem().String()
}

func TestNewMotionControlService(t *testing.T) {
	service := NewMotionControlService(Config{})
	if service == nil {
		t.Fatal("expected non-nil service")
	}
	if got := service.Source(); got != aivideokling.SourceWellAPIKlingMotionControl {
		t.Fatalf("unexpected source: %s", got)
	}
	if _, ok := service.(*providerwellapi.KlingMotionControlService); !ok {
		t.Fatalf("expected *wellapi.KlingMotionControlService, got %T", service)
	}
}

func TestNewEffectsService(t *testing.T) {
	service := NewEffectsService(Config{})
	if service == nil {
		t.Fatal("expected non-nil service")
	}
	if got := service.Source(); got != aivideokling.SourceWellAPIKlingEffects {
		t.Fatalf("unexpected source: %s", got)
	}
	if _, ok := service.(*providerwellapi.KlingEffectsService); !ok {
		t.Fatalf("expected *wellapi.KlingEffectsService, got %T", service)
	}
}

func TestConstructorsPreserveAPIKeyPath(t *testing.T) {
	motion := NewMotionControlService(Config{APIKey: "motion-key"})
	motionImpl, ok := motion.(*providerwellapi.KlingMotionControlService)
	if !ok {
		t.Fatalf("expected *wellapi.KlingMotionControlService, got %T", motion)
	}
	motionValue := reflect.ValueOf(motionImpl).Elem()
	taskValue := motionValue.FieldByName("task")
	clientValue := reflect.Indirect(taskValue).FieldByName("client")
	apiKeyValue := reflect.Indirect(clientValue).FieldByName("apiKey")
	if got := readStringField(t, apiKeyValue); got != "motion-key" {
		t.Fatalf("expected motion api key motion-key, got %q", got)
	}

	effects := NewEffectsService(Config{APIKey: "effects-key"})
	effectsImpl, ok := effects.(*providerwellapi.KlingEffectsService)
	if !ok {
		t.Fatalf("expected *wellapi.KlingEffectsService, got %T", effects)
	}
	effectsValue := reflect.ValueOf(effectsImpl).Elem()
	effectsTask := effectsValue.FieldByName("task")
	effectsClient := reflect.Indirect(effectsTask).FieldByName("client")
	effectsAPIKey := reflect.Indirect(effectsClient).FieldByName("apiKey")
	if got := readStringField(t, effectsAPIKey); got != "effects-key" {
		t.Fatalf("expected effects api key effects-key, got %q", got)
	}
}
