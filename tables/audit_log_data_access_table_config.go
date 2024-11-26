package tables

type AuditLogDataAccessTableConfig struct {
}

func (c *AuditLogDataAccessTableConfig) Validate() error {
	return nil
}

func (c *AuditLogDataAccessTableConfig) Identifier() string {
	return AuditLogDataAccessTableIdentifier
}
