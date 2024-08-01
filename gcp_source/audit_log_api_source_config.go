package gcp_source

type AuditLogAPISourceConfig struct {
	Credentials *string  `hcl:"credentials"`
	Project     string   `hcl:"project"`
	LogTypes    []string `hcl:"log_types"`
}

func (a *AuditLogAPISourceConfig) Validate() error {
	//TODO #graza implement me

	return nil
}
