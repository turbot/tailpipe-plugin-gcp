package tables

type ActivityAuditLogTableConfig struct {
}

func (c *ActivityAuditLogTableConfig) Validate() error {
	return nil
}

func (c *ActivityAuditLogTableConfig) Identifier() string {
	return ActivityAuditLogTableIdentifier
}
