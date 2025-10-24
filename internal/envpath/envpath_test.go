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

func TestSwitchWithCustomEnvironmentFromEnvVar(t *testing.T) {
	// Set custom environment via environment variable
	t.Setenv("CENV_ENVIRONMENTS", "prod-us-east-1 test-eu-central-1")

	root := t.TempDir()
	prodUSPath := filepath.Join(root, "prod-us-east-1", "services", "app")
	testEUPath := filepath.Join(root, "test-eu-central-1", "services", "app")
	mustMkdirAll(t, prodUSPath)

	target, err := envpath.Switch(prodUSPath, "test-eu-central-1")
	if err != nil {
		t.Fatalf("Switch returned error: %v", err)
	}

	if target != testEUPath {
		t.Fatalf("expected target %q, got %q", testEUPath, target)
	}
}

func TestSwitchWithCustomEnvironmentFromConfigFile(t *testing.T) {
	// Create a temporary home directory
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Create .cenvrc with custom environments
	cenvrcPath := filepath.Join(tmpHome, ".cenvrc")
	cenvrcContent := `# Custom environments
prod-us-east-1
test-eu-central-1  # Europe region
staging
`
	if err := os.WriteFile(cenvrcPath, []byte(cenvrcContent), 0o644); err != nil {
		t.Fatalf("failed to write .cenvrc: %v", err)
	}

	root := t.TempDir()
	stagingPath := filepath.Join(root, "staging", "services", "app")
	prodPath := filepath.Join(root, "prod-us-east-1", "services", "app")
	mustMkdirAll(t, stagingPath)

	target, err := envpath.Switch(stagingPath, "prod-us-east-1")
	if err != nil {
		t.Fatalf("Switch returned error: %v", err)
	}

	if target != prodPath {
		t.Fatalf("expected target %q, got %q", prodPath, target)
	}
}

func TestSwitchWithBothConfigFileAndEnvVar(t *testing.T) {
	// Create a temporary home directory
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Create .cenvrc
	cenvrcPath := filepath.Join(tmpHome, ".cenvrc")
	if err := os.WriteFile(cenvrcPath, []byte("staging\n"), 0o644); err != nil {
		t.Fatalf("failed to write .cenvrc: %v", err)
	}

	// Also set environment variable
	t.Setenv("CENV_ENVIRONMENTS", "canary")

	root := t.TempDir()
	stagingPath := filepath.Join(root, "staging", "services")
	canaryPath := filepath.Join(root, "canary", "services")
	mustMkdirAll(t, stagingPath)

	// Test switching from staging (config file) to canary (env var)
	target, err := envpath.Switch(stagingPath, "canary")
	if err != nil {
		t.Fatalf("Switch returned error: %v", err)
	}

	if target != canaryPath {
		t.Fatalf("expected target %q, got %q", canaryPath, target)
	}
}

func mustMkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("failed to create %q: %v", path, err)
	}
}
