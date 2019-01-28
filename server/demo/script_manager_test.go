package demo_test

import (
	"github.com/DSchalla/MatterDemo-Plugin/server/demo"
	"testing"
)

func TestNewScriptManager(t *testing.T) {
	sm, err := demo.NewScriptManager(nil,"../test/scripts")

	if err != nil {
		t.Errorf("Expected nil, got %s", err)
	}

	expectedCount := 3
	actualCount := sm.GetScriptCount()

	if expectedCount != actualCount {
		t.Errorf("Expected count %d, got %d", expectedCount, actualCount)
	}

	expectedScripts := []struct{
		Id string
		Name string
	}{
		{
			"incident_response",
			"Incident Response",
		},
	}

	for _, expectedScript := range expectedScripts {
		actualScript, err := sm.GetScript(expectedScript.Id)

		if err != nil {
			t.Errorf("Error fetching script with id %s: %s", expectedScript.Id, err)
			continue
		}

		if actualScript.Name != expectedScript.Name {
			t.Errorf("Expected %s, got %s", expectedScript.Name, actualScript.Name)
		}
	}
}