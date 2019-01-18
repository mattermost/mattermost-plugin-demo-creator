package demo_test

import (
	"github.com/DSchalla/MatterDemo-Plugin/server/demo"
	"testing"
)

func Test_LoadScriptsFromFile(t *testing.T) {
	filepath := "../test/demo_scripts.yml"
	scripts, err := demo.LoadScriptsFromFile(filepath)

	if err != nil {
		t.Errorf("Error loading sample file: %s", err)
	}

	expectedName := "Incident Response"
	if scripts[0].Name != expectedName {
		t.Errorf("Expected %s, actual %s", expectedName, scripts[0].Name)
	}
}