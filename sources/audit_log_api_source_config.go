package sources

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
)

type AuditLogAPISourceConfig struct {
	// required to allow partial decoding
	Remain hcl.Body `hcl:",remain" json:"-"`

	LogTypes []string `hcl:"log_types"`
}

func (a *AuditLogAPISourceConfig) Validate() error {
	if len(a.LogTypes) == 0 {
		return fmt.Errorf("log_types are required")
	}

	return nil
}

func (a *AuditLogAPISourceConfig) Identifier() string {
	return AuditLogAPISourceIdentifier
}
