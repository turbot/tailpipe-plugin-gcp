package tables

type AuditSystemEventLogTableConfig struct {
}

// Validate implements parse.Config.
func (a *AuditSystemEventLogTableConfig) Validate() error {
	panic("unimplemented")
}

func (a *AuditSystemEventLogTable) Validate() error {
	return nil
}
