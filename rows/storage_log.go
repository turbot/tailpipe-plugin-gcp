package rows

import (
	"time"

	"github.com/turbot/tailpipe-plugin-sdk/enrichment"
)

// StorageLog represents a structured log entry for GCP Storage logs.
type StorageLog struct {
    // embed required enrichment fields
	enrichment.CommonFields

	InsertID                 string   `json:"insertId"`
	Timestamp                time.Time   `json:"timestamp"`
	Severity                 string   `json:"severity"`
	LogName                  string   `json:"logName"`
	// ReceiveTimestamp         string   `json:"receiveTimestamp"`
	StatusCode               int32      `json:"statusCode,omitempty"`
	StatusMessage            string   `json:"statusMessage,omitempty"`
	// AuthenticationPrincipal  string   `json:"authenticationPrincipal"`
	// CallerIp                 string   `json:"callerIp"`
	// CallerSuppliedUserAgent  string   `json:"callerSuppliedUserAgent"`
	// RequestTime              time.Time   `json:"requestTime"`
	ServiceName              string   `json:"serviceName"`
	MethodName               string   `json:"methodName"`
	ResourceName             string   `json:"resourceName"`
	// ResourceCurrentLocations []string `json:"resourceCurrentLocations"`
	// Resource                 string   `json:"resource"`
	// Permission               string   `json:"permission"`
	// Location                 string   `json:"location"`
	// BucketName               string   `json:"bucketName"`
	// ProjectId                string   `json:"projectId"`
	// Granted                  bool     `json:"granted"`
	ResourceType             string   `json:"resourceType"`
}

