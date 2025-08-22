package logging_log_entry

import (
	"fmt"
	"slices"
	"strings"

	"github.com/hashicorp/hcl/v2"
)

type LoggingLogEntrySourceConfig struct {
	// required to allow partial decoding
	Remain   hcl.Body `hcl:",remain" json:"-"`
	LogTypes []string `hcl:"log_types,optional" json:"log_types"`
}

func (a *LoggingLogEntrySourceConfig) Validate() error {
	// Basic validation - just check if log types are provided
	// Table-specific validation will be done in the Init method
	return nil
}

// ValidateForTable validates log types for a specific table
func (a *LoggingLogEntrySourceConfig) ValidateForTable(tableName string) error {
	logTypeFilter := NewLogTypeFilter()

	if !logTypeFilter.IsValidTable(tableName) {
		return fmt.Errorf("invalid table name %s", tableName)
	}

	validLogTypes := logTypeFilter.GetAvailableLogTypesForTable(tableName)

	for _, logType := range a.LogTypes {
		if !slices.Contains(validLogTypes, logType) {
			return fmt.Errorf("invalid log type %s for table %s, valid log types are %s", logType, tableName, strings.Join(validLogTypes, ", "))
		}
	}
	return nil
}

func (a *LoggingLogEntrySourceConfig) Identifier() string {
	return LoggingLogEntrySourceIdentifier
}
