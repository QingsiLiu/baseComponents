package kling

import (
	"testing"

	aivideokling "github.com/QingsiLiu/baseComponents/service/aivideo/kling"
)

func TestCompatibilityLayerMatchesAIVideoKling(t *testing.T) {
	if SourceWellAPIKlingMotionControl != aivideokling.SourceWellAPIKlingMotionControl {
		t.Fatal("expected motion-control source to match aivideo/kling")
	}
	if ServiceTypeWellAPIKlingEffects != aivideokling.ServiceTypeWellAPIKlingEffects {
		t.Fatal("expected effects service type to match aivideo/kling")
	}
	if ConvertTaskStatus("succeed") != aivideokling.ConvertTaskStatus("succeed") {
		t.Fatal("expected status conversion to match aivideo/kling")
	}
	if got := GetServiceSource(ServiceTypeWellAPIKlingMotionControl); got != aivideokling.SourceWellAPIKlingMotionControl {
		t.Fatalf("unexpected source lookup result: %s", got)
	}
}
