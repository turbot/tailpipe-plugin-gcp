package tables

type AuditDataAccessLogTableConfig struct {
}

// Validate implements parse.Config.
func (a *AuditDataAccessLogTableConfig) Validate() error {
	panic("unimplemented")
}

func (a *AuditDataAccessLogTable) Validate() error {
	return nil
}
