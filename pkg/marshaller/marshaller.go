// Package marshaller
package marshaller

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/odlev/xray-cli/pkg/entity"
)

func Marshal(cfg *entity.Config, filePath string) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal to json: %w", err)
	}
	if err := os.WriteFile(filePath, data, 0o644); err != nil {
		return fmt.Errorf("failed write config: %w", err)
	}
	return nil
}
