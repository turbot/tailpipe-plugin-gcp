package sources

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
)

type AuditLogAPISourceConfig struct {
	// required to allow partial decoding
	Remain hcl.Body `hcl:",remain" json:"-"`

	//Credentials *string  `hcl:"credentials"`
	//Project     string   `hcl:"project"`
	LogTypes []string `hcl:"log_types"`
}

func (a *AuditLogAPISourceConfig) Validate() error {
	//if a.Project == "" {
	//	return fmt.Errorf("project is required")
	//}
	if len(a.LogTypes) == 0 {
		return fmt.Errorf("log_types are required")
	}
	//if a.Credentials == nil {
	//	return fmt.Errorf("credentials are required")
	//}

	return nil
}

func (a *AuditLogAPISourceConfig) Identifier() string {
	return AuditLogAPISourceIdentifier
}
