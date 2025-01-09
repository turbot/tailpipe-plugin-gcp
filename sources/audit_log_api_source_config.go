package sources

import (
	"github.com/hashicorp/hcl/v2"
)

type AuditLogAPISourceConfig struct {
	// required to allow partial decoding
	Remain   hcl.Body `hcl:",remain" json:"-"`
	LogTypes []string `hcl:"log_types,optional" json:"log_types"`
}

func (a *AuditLogAPISourceConfig) Validate() error {
	return nil
}

func (a *AuditLogAPISourceConfig) Identifier() string {
	return AuditLogAPISourceIdentifier
}
