package tests

import (
	"neawn-backend/internal/app"
	"testing"
)

func TestApp(t *testing.T) {
	app := app.New()

	if app == nil {
		t.Error("Expected app to not be nil")
	}
}
