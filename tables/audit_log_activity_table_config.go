package tables

type AuditLogActivityTableConfig struct {
}

func (c *AuditLogActivityTableConfig) Validate() error {
	return nil
}

func (c *AuditLogActivityTableConfig) Identifier() string {
	return AuditLogActivityTableIdentifier
}
