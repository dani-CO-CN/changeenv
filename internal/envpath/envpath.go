package envpath

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

var defaultEnvs = []string{"dev", "test", "prod"}

// Switch returns the equivalent path in the target environment.
//
// fromPath must point to a directory that sits under an environment directory
// (for example ".../dev/..."). The first environment segment encountered in the
// path is replaced with targetEnv. Known environment names are dev, test, and
// prod.
func Switch(fromPath, targetEnv string) (string, error) {
	targetEnv = strings.TrimSpace(targetEnv)
	if targetEnv == "" {
		return "", errors.New("target environment must not be empty")
	}
	if fromPath == "" {
		return "", errors.New("current path must not be empty")
	}

	cleanFrom := filepath.Clean(fromPath)
	volume, hasLeading, parts := splitPath(cleanFrom)
	if len(parts) == 0 {
		return "", fmt.Errorf("path %q is not inside a known environment", fromPath)
	}

	envIndex := -1
	for idx, part := range parts {
		if strings.EqualFold(part, targetEnv) || isKnownEnv(part) {
			envIndex = idx
			break
		}
	}
	if envIndex == -1 {
		return "", fmt.Errorf("path %q is not inside a known environment", fromPath)
	}

	parts[envIndex] = targetEnv

	targetPath := assemblePath(volume, hasLeading, parts)

	return targetPath, nil
}

func isKnownEnv(name string) bool {
	for _, env := range defaultEnvs {
		if strings.EqualFold(env, name) {
			return true
		}
	}
	return false
}

func splitPath(path string) (volume string, hasLeading bool, parts []string) {
	volume = filepath.VolumeName(path)
	rest := path[len(volume):]
	if rest == "" {
		return volume, false, nil
	}
	seps := string(filepath.Separator)
	hasLeading = strings.HasPrefix(rest, seps)
	slashed := filepath.ToSlash(rest)
	slashed = strings.TrimPrefix(slashed, "/")
	if slashed == "" {
		return volume, hasLeading, nil
	}
	parts = strings.Split(slashed, "/")
	return volume, hasLeading, parts
}

func assemblePath(volume string, hasLeading bool, parts []string) string {
	joined := filepath.Join(parts...)
	if hasLeading && joined != "" {
		joined = string(filepath.Separator) + joined
	}
	if volume != "" {
		joined = volume + joined
	}
	if joined == "" {
		if volume != "" {
			return volume + string(filepath.Separator)
		}
		return string(filepath.Separator)
	}
	return joined
}
