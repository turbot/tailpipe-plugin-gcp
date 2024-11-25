package tables

type DataAccessAuditLogTableConfig struct {
}

func (c *DataAccessAuditLogTableConfig) Validate() error {
	return nil
}

func (c *DataAccessAuditLogTableConfig) Identifier() string {
	return DataAccessAuditLogTableIdentifier
}
