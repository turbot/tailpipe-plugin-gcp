package audit_log_api

import (
	"fmt"
	"slices"
	"strings"

	"github.com/hashicorp/hcl/v2"
)

type AuditLogAPISourceConfig struct {
	// required to allow partial decoding
	Remain   hcl.Body `hcl:",remain" json:"-"`
	LogTypes []string `hcl:"log_types,optional" json:"log_types"`
}

func (a *AuditLogAPISourceConfig) Validate() error {
	validLogTypes := []string{"activity", "data_access", "system_event", "policy", "cloud_run_request", "app_engine_request", "requests"}

	for _, logType := range a.LogTypes {
		if !slices.Contains(validLogTypes, logType) {
			return fmt.Errorf("invalid log type %s, valid log types are %s", logType, strings.Join(validLogTypes, ", "))
		}
	}
	return nil
}

func (a *AuditLogAPISourceConfig) Identifier() string {
	return AuditLogAPISourceIdentifier
}
