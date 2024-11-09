package sources

import "fmt"

type AuditLogAPISourceConfig struct {
	Credentials *string `hcl:"credentials"`
	Project     string  `hcl:"project"`
	LogType     string  `hcl:"log_type"`
}

func (a *AuditLogAPISourceConfig) Validate() error {
	if a.Project == "" {
		return fmt.Errorf("project is required")
	}
	if a.LogType == "" {
		return fmt.Errorf("log type is required")
	}
	if a.Credentials == nil {
		return fmt.Errorf("credentials is required")
	}

	return nil
}
