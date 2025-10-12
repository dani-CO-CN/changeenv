package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const shellHelperSnippet = `cenv() { cd "$(changeenv "$1")"; }`

func runConfigure(autoApply bool) error {
	shellName := detectShellName(os.Getenv("SHELL"))
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("determine user home directory: %w", err)
	}

	binaryDir, execErr := executableDir()
	if execErr != nil {
		binaryDir = ""
	} else if absDir, err := filepath.Abs(binaryDir); err == nil {
		binaryDir = absDir
	}
	onPath := dirOnPath(binaryDir, os.Getenv("PATH"), homeDir)

	configPath := resolveConfigPath(shellName, homeDir, os.Getenv("ZDOTDIR"), fileExists)
	if configPath == "" {
		printConfigureInstructions(shellName, "", autoApply, binaryDir, onPath)
		if autoApply {
			return errors.New("unable to determine configuration file path; rerun without --create")
		}
		return nil
	}

	printConfigureInstructions(shellName, displayPath(configPath, homeDir), autoApply, binaryDir, onPath)

	if !autoApply {
		return nil
	}

	applied, err := appendSnippet(configPath, shellHelperSnippet)
	if err != nil {
		return err
	}
	if applied {
		fmt.Fprintf(os.Stdout, "Added helper function to %s\n", displayPath(configPath, homeDir))
	} else {
		fmt.Fprintf(os.Stdout, "Helper function already present in %s\n", displayPath(configPath, homeDir))
	}

	if binaryDir == "" {
		return nil
	}
	if onPath {
		fmt.Fprintf(os.Stdout, "Binary directory already present in PATH\n")
		return nil
	}

	pathSnippet := exportPathSnippet(binaryDir)
	addedPath, err := appendSnippet(configPath, pathSnippet)
	if err != nil {
		return err
	}
	if addedPath {
		fmt.Fprintf(os.Stdout, "Added binary directory to PATH in %s\n", displayPath(configPath, homeDir))
	} else {
		fmt.Fprintf(os.Stdout, "Binary directory already exported in %s\n", displayPath(configPath, homeDir))
	}

	return nil
}

func detectShellName(shellPath string) string {
	if shellPath == "" {
		return ""
	}
	return filepath.Base(shellPath)
}

func resolveConfigPath(shellName, homeDir, zdotdir string, exists func(string) bool) string {
	var candidates []string
	addCandidate := func(path string) {
		if path == "" {
			return
		}
		for _, candidate := range candidates {
			if candidate == path {
				return
			}
		}
		candidates = append(candidates, path)
	}

	switch shellName {
	case "zsh":
		if zdotdir != "" {
			addCandidate(filepath.Join(zdotdir, ".zshrc"))
		}
		if homeDir != "" {
			addCandidate(filepath.Join(homeDir, ".zshrc"))
		}
	case "bash":
		if homeDir != "" {
			addCandidate(filepath.Join(homeDir, ".bashrc"))
			addCandidate(filepath.Join(homeDir, ".bash_profile"))
		}
	}
	if homeDir != "" {
		addCandidate(filepath.Join(homeDir, ".profile"))
	}

	for _, candidate := range candidates {
		if exists(candidate) {
			return candidate
		}
	}
	if len(candidates) == 0 {
		return ""
	}
	return candidates[0]
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func appendSnippet(configPath, snippet string) (bool, error) {
	data, err := os.ReadFile(configPath)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return false, fmt.Errorf("read %s: %w", configPath, err)
	}

	if bytes.Contains(data, []byte(snippet)) {
		return false, nil
	}

	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		return false, fmt.Errorf("create config directory: %w", err)
	}

	f, err := os.OpenFile(configPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return false, fmt.Errorf("open %s: %w", configPath, err)
	}
	defer f.Close()

	var toWrite strings.Builder
	if len(data) > 0 && data[len(data)-1] != '\n' {
		toWrite.WriteString("\n")
	}
	toWrite.WriteString("\n")
	toWrite.WriteString(snippet)
	toWrite.WriteString("\n")

	if _, err := f.WriteString(toWrite.String()); err != nil {
		return false, fmt.Errorf("write %s: %w", configPath, err)
	}
	return true, nil
}

func printConfigureInstructions(shellName, configDisplayPath string, autoApply bool, binaryDir string, onPath bool) {
	if shellName == "" {
		fmt.Fprintln(os.Stdout, "Could not determine the active shell from $SHELL.")
	} else {
		fmt.Fprintf(os.Stdout, "Detected shell: %s\n", shellName)
	}

	fmt.Fprintln(os.Stdout)
	fmt.Fprintln(os.Stdout, "Add the following function to your shell configuration:")
	fmt.Fprintf(os.Stdout, "\n%s\n\n", shellHelperSnippet)

	if configDisplayPath != "" {
		fmt.Fprintf(os.Stdout, "Suggested command:\n  printf '\\n%%s\\n' %s >> %s\n", shellSingleQuote(shellHelperSnippet), configDisplayPath)
		if binaryDir != "" && !onPath {
			fmt.Fprintf(os.Stdout, "Add the binary directory to PATH:\n  printf '\\n%%s\\n' %s >> %s\n", shellSingleQuote(exportPathSnippet(binaryDir)), configDisplayPath)
		} else if binaryDir == "" {
			fmt.Fprintf(os.Stdout, "Ensure the directory containing changeenv is on your PATH.\n")
		}
		if !autoApply {
			fmt.Fprintf(os.Stdout, "Or run: changeenv configure --create\n")
		}
	} else {
		fmt.Fprintln(os.Stdout, "Unable to determine a configuration file path automatically.")
	}
}

func displayPath(path, homeDir string) string {
	if homeDir != "" && strings.HasPrefix(path, homeDir+string(filepath.Separator)) {
		return filepath.Join("~", strings.TrimPrefix(path[len(homeDir):], string(filepath.Separator)))
	}
	if homeDir != "" && path == homeDir {
		return "~"
	}
	return path
}

func executableDir() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}
	if exePath == "" {
		return "", errors.New("executable path is empty")
	}
	if resolved, err := filepath.EvalSymlinks(exePath); err == nil && resolved != "" {
		exePath = resolved
	}
	return filepath.Dir(exePath), nil
}

func dirOnPath(dir, pathEnv, homeDir string) bool {
	if dir == "" {
		return true
	}
	cleanDir := filepath.Clean(dir)
	if absDir, err := filepath.Abs(cleanDir); err == nil {
		cleanDir = absDir
	}
	entries := filepath.SplitList(pathEnv)
	for _, entry := range entries {
		if entry == "" {
			continue
		}
		entry = expandTilde(entry, homeDir)
		cleanEntry := filepath.Clean(entry)
		if absEntry, err := filepath.Abs(cleanEntry); err == nil {
			cleanEntry = absEntry
		}
		if pathsEqual(cleanEntry, cleanDir) {
			return true
		}
	}
	return false
}

func pathsEqual(a, b string) bool {
	if runtime.GOOS == "windows" {
		return strings.EqualFold(a, b)
	}
	return a == b
}

func exportPathSnippet(dir string) string {
	return fmt.Sprintf(`export PATH="$PATH:%s"`, escapeForDoubleQuotes(dir))
}

func escapeForDoubleQuotes(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `"`, `\"`)
	return value
}

func shellSingleQuote(value string) string {
	return "'" + strings.ReplaceAll(value, `'`, `'\''`) + "'"
}

func expandTilde(path, homeDir string) string {
	if homeDir == "" {
		return path
	}
	if path == "~" {
		return homeDir
	}
	if strings.HasPrefix(path, "~"+string(filepath.Separator)) {
		return filepath.Join(homeDir, path[2:])
	}
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(homeDir, path[2:])
	}
	return path
}
