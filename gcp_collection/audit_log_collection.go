package gcp_collection

import (
	"context"
	"fmt"
	"github.com/rs/xid"
	"github.com/turbot/tailpipe-plugin-sdk/helpers"
	"time"

	"cloud.google.com/go/logging"
	"github.com/turbot/tailpipe-plugin-gcp/gcp_source"
	"github.com/turbot/tailpipe-plugin-gcp/gcp_types"
	"github.com/turbot/tailpipe-plugin-sdk/collection"
	"github.com/turbot/tailpipe-plugin-sdk/enrichment"
	"github.com/turbot/tailpipe-plugin-sdk/plugin"
)

type AuditLogCollection struct {
	collection.Base

	Config *gcp_types.AuditLogCollectionConfig
}

func NewAuditLogCollection() plugin.Collection {
	return &AuditLogCollection{}
}

func (c *AuditLogCollection) Identifier() string {
	return "gcp_audit_log"
}

func (c *AuditLogCollection) GetRowSchema() any {
	return gcp_types.AuditLogRow{}
}

func (c *AuditLogCollection) GetConfigSchema() any {
	return &gcp_types.AuditLogCollectionConfig{}
}

func (c *AuditLogCollection) Init(ctx context.Context, configData []byte) error {
	fmt.Println("GCP Collection Init") // TODO: #debug remove

	// TODO: #config use actual configuration (& validate, etc)
	tmpPath := "~/gcp/tailpipe-creds.json"
	config := &gcp_types.AuditLogCollectionConfig{
		Credentials: &tmpPath,
	}

	c.Config = config

	// TODO: #config create source from config
	source := gcp_source.NewAuditLogAPISource(config)
	return c.AddSource(source)
}

func (c *AuditLogCollection) EnrichRow(row any, sourceEnrichmentFields *enrichment.CommonFields) (any, error) {
	item, ok := row.(logging.Entry)
	if !ok {
		return nil, fmt.Errorf("invalid row type: %T, expected logging.Entry", row)
	}

	// TODO: Validate sourceEnrichmentFields

	record := &gcp_types.AuditLogRow{
		CommonFields: *sourceEnrichmentFields,
		Entry:        item,
	}

	// Record standardization
	record.TpID = xid.New().String()
	record.TpSourceType = "gcp_audit_log" // TODO: Verify (might be able to use specific log type)
	record.TpTimestamp = helpers.UnixMillis(item.Timestamp.UnixNano() / int64(time.Millisecond))
	record.TpIngestTimestamp = helpers.UnixMillis(time.Now().UnixNano() / int64(time.Millisecond))

	// TODO: Figure out other mappings

	// Hive Fields
	record.TpCollection = "gcp_audit_log"
	record.TpYear = int32(item.Timestamp.Year())
	record.TpMonth = int32(item.Timestamp.Month())
	record.TpDay = int32(item.Timestamp.Day())

	return record, nil
}
