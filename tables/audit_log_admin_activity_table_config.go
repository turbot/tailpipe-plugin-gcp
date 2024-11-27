package tables

type AuditLogAdminActivityTableConfig struct {
}

func (c *AuditLogAdminActivityTableConfig) Validate() error {
	return nil
}

func (c *AuditLogAdminActivityTableConfig) Identifier() string {
	return AuditLogAdminActivityTableIdentifier
}
