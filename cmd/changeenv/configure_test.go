package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveConfigPathPrefersZdotdir(t *testing.T) {
	home := t.TempDir()
	zdotdir := filepath.Join(home, "custom")
	expected := filepath.Join(zdotdir, ".zshrc")

	exists := func(path string) bool {
		return path == expected
	}

	got := resolveConfigPath("zsh", home, zdotdir, exists)
	if got != expected {
		t.Fatalf("expected config path %q, got %q", expected, got)
	}
}

func TestResolveConfigPathFallsBackToProfile(t *testing.T) {
	home := t.TempDir()
	expected := filepath.Join(home, ".profile")

	got := resolveConfigPath("fish", home, "", func(string) bool { return false })
	if got != expected {
		t.Fatalf("expected fallback %q, got %q", expected, got)
	}
}

func TestResolveConfigPathPrefersExistingBashrc(t *testing.T) {
	home := t.TempDir()
	bashrc := filepath.Join(home, ".bashrc")

	exists := func(path string) bool {
		return path == bashrc
	}

	got := resolveConfigPath("bash", home, "", exists)
	if got != bashrc {
		t.Fatalf("expected %q, got %q", bashrc, got)
	}
}

func TestAppendSnippetCreatesAndAppends(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".zshrc")

	applied, err := appendSnippet(configPath, shellHelperSnippet)
	if err != nil {
		t.Fatalf("appendSnippet returned error: %v", err)
	}
	if !applied {
		t.Fatalf("expected helper to be appended")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config path: %v", err)
	}
	want := "\n" + shellHelperSnippet + "\n"
	if string(data) != want {
		t.Fatalf("unexpected config contents %q", string(data))
	}
}

func TestAppendSnippetSkipsDuplicates(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".zshrc")

	if err := os.WriteFile(configPath, []byte(shellHelperSnippet+"\n"), 0o644); err != nil {
		t.Fatalf("write initial config: %v", err)
	}

	applied, err := appendSnippet(configPath, shellHelperSnippet)
	if err != nil {
		t.Fatalf("appendSnippet returned error: %v", err)
	}
	if applied {
		t.Fatalf("expected no changes when helper already present")
	}
}

func TestDisplayPathUsesTilde(t *testing.T) {
	home := "/Users/example"
	path := filepath.Join(home, ".zshrc")
	got := displayPath(path, home)
	expected := filepath.Join("~", ".zshrc")
	if got != expected {
		t.Fatalf("expected %q, got %q", expected, got)
	}
}

func TestDirOnPathMatchesAbsoluteEntry(t *testing.T) {
	dir := t.TempDir()
	pathEnv := strings.Join([]string{"/usr/bin", dir}, string(os.PathListSeparator))

	if !dirOnPath(dir, pathEnv, "") {
		t.Fatalf("expected dirOnPath to find %q in %q", dir, pathEnv)
	}
}

func TestDirOnPathHandlesMissingEntry(t *testing.T) {
	dir := t.TempDir()
	pathEnv := strings.Join([]string{"/usr/bin", "/bin"}, string(os.PathListSeparator))

	if dirOnPath(dir, pathEnv, "") {
		t.Fatalf("expected dirOnPath to return false")
	}
}

func TestExportPathSnippetEscapesCharacters(t *testing.T) {
	dir := `/tmp/some "dir"/with\spaces`
	got := exportPathSnippet(dir)
	expected := `export PATH="$PATH:/tmp/some \"dir\"/with\\spaces"`
	if got != expected {
		t.Fatalf("expected %q, got %q", expected, got)
	}
}

func TestShellSingleQuoteEscapes(t *testing.T) {
	input := "need 'quotes'"
	got := shellSingleQuote(input)
	expected := `'need '\''quotes'\'''`
	if got != expected {
		t.Fatalf("expected %q, got %q", expected, got)
	}
}

func TestDirOnPathExpandsTilde(t *testing.T) {
	home := t.TempDir()
	dir := filepath.Join(home, "bin")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("create bin dir: %v", err)
	}
	pathEnv := strings.Join([]string{"~/bin"}, string(os.PathListSeparator))

	if !dirOnPath(dir, pathEnv, home) {
		t.Fatalf("expected dirOnPath to expand tilde")
	}
}
