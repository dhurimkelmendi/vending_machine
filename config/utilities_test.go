package config

import (
	"testing"
)

func dummyGetter(name string) string {
	if name == "VAR1" {
		return "one"
	}

	if name == "VAR2" {
		return "two"
	}

	if name == "BOOL1" {
		return "true"
	}

	return ""
}

func TestConfigUtilities(t *testing.T) {
	t.Parallel()

	appConfig := envConfigGetter(dummyGetter, true)

	t.Run("get single config", func(t *testing.T) {
		value := appConfig.GetConfig("VAR1", "")
		if value != "one" {
			t.Fatalf("Incorrect value: %s", value)
		}
	})

	t.Run("get single config with default", func(t *testing.T) {
		value := appConfig.GetConfig("VARX", "x")
		if value != "x" {
			t.Fatalf("Incorrect value: %s", value)
		}
	})

	t.Run("get single bool config", func(t *testing.T) {
		value := appConfig.GetFlag("BOOL1", false)
		if value != true {
			t.Fatalf("Incorrect value: %v", value)
		}
	})
}
