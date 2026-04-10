package serviceregistry

import "testing"

func TestRegistryMappings(t *testing.T) {
	registry := New([]Config{
		{Source: "owner", ServiceType: "0"},
		{Source: "foo", ServiceType: "f1"},
		{Source: "bar", ServiceType: "b2"},
	})

	if got := registry.GetServiceType("foo"); got != "f1" {
		t.Fatalf("expected foo -> f1, got %s", got)
	}
	if got := registry.GetServiceSource("b2"); got != "bar" {
		t.Fatalf("expected b2 -> bar, got %s", got)
	}
	if registry.GetServiceType("missing") != "unknown" {
		t.Fatal("expected missing source to return unknown")
	}
	if !registry.IsValidSource("bar") || !registry.IsValidServiceType("0") {
		t.Fatal("expected known source/type to be valid")
	}
	if registry.IsValidSource("missing") || registry.IsValidServiceType("missing") {
		t.Fatal("expected unknown source/type to be invalid")
	}
}

func TestRegistryReturnsCopies(t *testing.T) {
	registry := New([]Config{
		{Source: "foo", ServiceType: "f1"},
	})

	sources := registry.AllServiceSource()
	types := registry.AllServiceType()
	sources[0] = "mutated"
	types[0] = "mutated"

	if registry.AllServiceSource()[0] != "foo" {
		t.Fatal("expected source slice to be copied")
	}
	if registry.AllServiceType()[0] != "f1" {
		t.Fatal("expected type slice to be copied")
	}
}
