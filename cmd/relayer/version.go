package main

import (
	"fmt"
	"runtime/debug"
	"strings"
)

// GetVersion returns our version string for an executable
func getVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown-version"
	}

	// info.Main.Version is " (devel)" if no tag was passed to go install
	version := strings.Replace(info.Main.Version, "(devel)", "dev", 1)
	dirty := false

	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			// limit hash to 7 bytes (what "git rev-parse --short HEAD" returns)
			version = fmt.Sprintf("%s-%.7s", version, setting.Value)
		case "vcs.modified":
			if setting.Value == "true" {
				dirty = true
			}
		}
	}

	if dirty {
		version += "-dirty"
	}

	return fmt.Sprintf("%s-%s", version, info.GoVersion)
}
