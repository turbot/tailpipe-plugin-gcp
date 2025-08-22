package logging_log_entry

import "fmt"

// LogTypeFilter represents the mapping of table names to their log type filters
type LogTypeFilter struct {
	// TableLogTypeMap maps table name to a map of log type to filter values
	TableLogTypeMap map[string]map[string][]string
}

// NewLogTypeFilter creates a new LogTypeFilter with predefined mappings
// LogTypeFilter defines the mapping between Steampipe tables and the supported
// Google Cloud log types.
//
// ⚠️ IMPORTANT: If you add a new table here, make sure to configure its log
// type keys according to the naming conventions shown in the Google Cloud
// Console (Logs Explorer > "Select log names" dropdown).
//
// Example for the table gcp_audit_log (from the console):
//   - activity        → cloudaudit.googleapis.com/activity
//   - data_access     → cloudaudit.googleapis.com/data_access
//   - system_event    → cloudaudit.googleapis.com/system_event
//   - policy          → cloudaudit.googleapis.com/policy
//
// The values in TableLogTypeMap should be fully qualified log resource names
// in the form `projects/%s/logs/<log_id>`, where <log_id> matches the console.
// Be careful to URL-encode path separators as needed (e.g., `%%2F` for “/”).
//
// This ensures that queries made via Steampipe align with what users see in
// the GCP Logs Explorer UI.
func NewLogTypeFilter() *LogTypeFilter {
	return &LogTypeFilter{
		TableLogTypeMap: map[string]map[string][]string{
			"gcp_audit_log": {
				"activity":     []string{"projects/%s/logs/cloudaudit.googleapis.com%%2Factivity"},
				"data_access":  []string{"projects/%s/logs/cloudaudit.googleapis.com%%2Fdata_access"},
				"system_event": []string{"projects/%s/logs/cloudaudit.googleapis.com%%2Fsystem_event"},
				"policy":       []string{"projects/%s/logs/cloudaudit.googleapis.com%%2Fpolicy"},
			},
		},
	}
}

// GetLogFiltersForTable returns the log filters for a specific table and project
func (l *LogTypeFilter) GetLogFiltersForTable(tableName, projectID string) []string {
	if tableConfig, exists := l.TableLogTypeMap[tableName]; exists {
		var filters []string
		for _, filterTemplates := range tableConfig {
			for _, template := range filterTemplates {
				filter := fmt.Sprintf(template, projectID)
				filters = append(filters, filter)
			}
		}
		return filters
	}
	return nil
}

// GetLogFiltersForTableAndType returns the log filters for a specific table, log type, and project
func (l *LogTypeFilter) GetLogFiltersForTableAndType(tableName, logType, projectID string) []string {
	if tableConfig, exists := l.TableLogTypeMap[tableName]; exists {
		if filterTemplates, exists := tableConfig[logType]; exists {
			var filters []string
			for _, template := range filterTemplates {
				filter := fmt.Sprintf(template, projectID)
				filters = append(filters, filter)
			}
			return filters
		}
	}
	return nil
}

// GetAvailableLogTypesForTable returns all available log types for a specific table
func (l *LogTypeFilter) GetAvailableLogTypesForTable(tableName string) []string {
	if tableConfig, exists := l.TableLogTypeMap[tableName]; exists {
		var logTypes []string
		for logType := range tableConfig {
			logTypes = append(logTypes, logType)
		}
		return logTypes
	}
	return nil
}

// IsValidTable returns true if the table name is valid
func (l *LogTypeFilter) IsValidTable(tableName string) bool {
	_, exists := l.TableLogTypeMap[tableName]
	return exists
}

// IsValidLogTypeForTable returns true if the log type is valid for the given table
func (l *LogTypeFilter) IsValidLogTypeForTable(tableName, logType string) bool {
	if tableConfig, exists := l.TableLogTypeMap[tableName]; exists {
		_, exists := tableConfig[logType]
		return exists
	}
	return false
}
