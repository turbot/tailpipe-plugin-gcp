package audit_log_api

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/hashicorp/hcl/v2"
)

type AuditLogAPISourceConfig struct {
	// required to allow partial decoding
	Remain   hcl.Body `hcl:",remain" json:"-"`
	LogTypes []string `hcl:"log_types,optional" json:"log_types"`

	// Retry configuration
	MaxRetries     *int           `hcl:"max_retries,optional" json:"max_retries"`
	InitialBackoff *time.Duration `hcl:"initial_backoff,optional" json:"initial_backoff"`
	MaxBackoff     *time.Duration `hcl:"max_backoff,optional" json:"max_backoff"`
}

func (a *AuditLogAPISourceConfig) Validate() error {
	validLogTypes := []string{"activity", "data_access", "system_event", "policy"}

	for _, logType := range a.LogTypes {
		if !slices.Contains(validLogTypes, logType) {
			return fmt.Errorf("invalid log type %s, valid log types are %s", logType, strings.Join(validLogTypes, ", "))
		}
	}

	// Validate retry configuration
	if a.MaxRetries != nil && *a.MaxRetries < 0 {
		return fmt.Errorf("max_retries must be non-negative, got %d", *a.MaxRetries)
	}

	if a.InitialBackoff != nil && *a.InitialBackoff < 0 {
		return fmt.Errorf("initial_backoff must be non-negative, got %v", *a.InitialBackoff)
	}

	if a.MaxBackoff != nil && *a.MaxBackoff < 0 {
		return fmt.Errorf("max_backoff must be non-negative, got %v", *a.MaxBackoff)
	}

	if a.InitialBackoff != nil && a.MaxBackoff != nil && *a.InitialBackoff > *a.MaxBackoff {
		return fmt.Errorf("initial_backoff (%v) cannot be greater than max_backoff (%v)", *a.InitialBackoff, *a.MaxBackoff)
	}

	return nil
}

// GetMaxRetries returns the configured max retries or a default value
func (a *AuditLogAPISourceConfig) GetMaxRetries() int {
	if a.MaxRetries != nil {
		return *a.MaxRetries
	}
	return 3 // default to 3 retries
}

// GetInitialBackoff returns the configured initial backoff or a default value
func (a *AuditLogAPISourceConfig) GetInitialBackoff() time.Duration {
	if a.InitialBackoff != nil {
		return *a.InitialBackoff
	}
	return 1 * time.Second // default to 1 second
}

// GetMaxBackoff returns the configured max backoff or a default value
func (a *AuditLogAPISourceConfig) GetMaxBackoff() time.Duration {
	if a.MaxBackoff != nil {
		return *a.MaxBackoff
	}
	return 30 * time.Second // default to 30 seconds
}

func (a *AuditLogAPISourceConfig) Identifier() string {
	return AuditLogAPISourceIdentifier
}
