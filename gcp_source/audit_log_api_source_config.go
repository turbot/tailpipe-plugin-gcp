package gcp_source

type AuditLogAPISourceConfig struct {
	Credentials *string
	Project     string
	LogTypes    []string
}

func (a AuditLogAPISourceConfig) Validate() error {
	//TODO #graza implement me
	return nil
}
