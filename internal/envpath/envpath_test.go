package envpath_test

import (
	"os"
	"path/filepath"
	"testing"

	"envchanger/internal/envpath"
)

func TestSwitchReturnsTargetPath(t *testing.T) {
	root := t.TempDir()
	devPath := filepath.Join(root, "dev", "eu-west6", "infra")
	prodPath := filepath.Join(root, "prod", "eu-west6", "infra")

	mustMkdirAll(t, devPath)
	mustMkdirAll(t, prodPath)

	target, err := envpath.Switch(devPath, "prod")
	if err != nil {
		t.Fatalf("Switch returned error: %v", err)
	}

	if target != prodPath {
		t.Fatalf("expected target %q, got %q", prodPath, target)
	}
}

func TestSwitchValidatesEnvironmentSegment(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "misc", "eu-west6")
	mustMkdirAll(t, path)

	_, err := envpath.Switch(path, "prod")
	if err == nil {
		t.Fatalf("expected error when no environment segment is present")
	}
}

func TestSwitchReturnsTargetEvenIfItDoesNotExist(t *testing.T) {
	root := t.TempDir()
	devPath := filepath.Join(root, "dev", "cluster")
	mustMkdirAll(t, devPath)

	target, err := envpath.Switch(devPath, "prod")
	if err != nil {
		t.Fatalf("Switch returned error: %v", err)
	}
	expected := filepath.Join(root, "prod", "cluster")
	if target != expected {
		t.Fatalf("expected %q, got %q", expected, target)
	}
}

func mustMkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("failed to create %q: %v", path, err)
	}
}
