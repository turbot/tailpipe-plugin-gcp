package tables

type AuditActivityLogTableConfig struct {
}

// Validate implements parse.Config.
func (a *AuditActivityLogTableConfig) Validate() error {
	panic("unimplemented")
}

func (a *AuditActivityLogTable) Validate() error {
	return nil
}
