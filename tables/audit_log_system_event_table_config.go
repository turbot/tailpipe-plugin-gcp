package tables

type AuditLogSystemEventTableConfig struct {
}

func (c *AuditLogSystemEventTableConfig) Validate() error {
	return nil
}

func (c *AuditLogSystemEventTableConfig) Identifier() string {
	return AuditLogSystemEventTableIdentifier
}
