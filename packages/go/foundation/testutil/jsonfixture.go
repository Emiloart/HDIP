package testutil

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

func AssertJSONMatchesFixture(t *testing.T, body []byte, fixturePath string) {
	t.Helper()

	expectedFixture, err := LoadJSONFixture(fixturePath)
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}

	var actual any
	if err := json.Unmarshal(body, &actual); err != nil {
		t.Fatalf("unmarshal actual body: %v", err)
	}

	if !reflect.DeepEqual(actual, expectedFixture) {
		t.Fatalf("unexpected json body\nexpected: %#v\nactual: %#v", expectedFixture, actual)
	}
}

func LoadJSONFixture(relativePath string) (any, error) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return nil, os.ErrNotExist
	}

	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", "..", "..", ".."))
	absolutePath := filepath.Join(repoRoot, relativePath)

	raw, err := os.ReadFile(absolutePath)
	if err != nil {
		return nil, err
	}

	var payload any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, err
	}

	return payload, nil
}
