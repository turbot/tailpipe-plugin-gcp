package tables

type SystemEventAuditLogTableConfig struct {
}

func (c *SystemEventAuditLogTableConfig) Validate() error {
	return nil
}

func (c *SystemEventAuditLogTableConfig) Identifier() string {
	return SystemEventAuditLogTableIdentifier
}
