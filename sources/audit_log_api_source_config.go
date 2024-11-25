package sources

import (
	"github.com/hashicorp/hcl/v2"
)

type AuditLogAPISourceConfig struct {
	// required to allow partial decoding
	Remain hcl.Body `hcl:",remain" json:"-"`
}

func (a *AuditLogAPISourceConfig) Validate() error {
	return nil
}

func (a *AuditLogAPISourceConfig) Identifier() string {
	return AuditLogAPISourceIdentifier
}
