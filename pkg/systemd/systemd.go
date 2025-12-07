// Package systemd provides helpers to emit systemd unit files.
package systemd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// WriteUnit writes a systemd unit to the provided file path.
func WriteUnit(systemdUnitPath, description, workingDir, execStart, restart, wantedBy string) error {
	absPath, err := filepath.Abs(systemdUnitPath)
	if err != nil {
		return fmt.Errorf("invalid service path: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		return fmt.Errorf("failed to prepare service directory: %w", err)
	}
	unit := fmt.Sprintf(`[Unit]
Description=%s
After=network.target

[Service]
Type=simple
WorkingDirectory=%s
ExecStart=%s
Restart=%s

[Install]
WantedBy=%s
`, description, workingDir, execStart, restart, wantedBy)
	if err := os.WriteFile(absPath, []byte(unit), 0o644); err != nil {
		return fmt.Errorf("failed to write service file: %w", err)
	}
	return nil
}

// EscapedExecStart builds an ExecStart value that properly quotes arguments.
func EscapedExecStart(binary, config string) string {
	return fmt.Sprintf("%s run -c %s", escapeArg(binary), escapeArg(config))
}

func escapeArg(value string) string {
	if value == "" {
		return "\"\""
	}
	if strings.ContainsAny(value, " \t'\"`()") {
		return strconv.Quote(value)
	}
	return value
}
