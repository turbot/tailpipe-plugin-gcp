package tables

import (
	"context"
	"time"

	"github.com/rs/xid"
	"github.com/turbot/tailpipe-plugin-gcp/config"
	"github.com/turbot/tailpipe-plugin-gcp/mappers"
	"github.com/turbot/tailpipe-plugin-gcp/rows"
	"github.com/turbot/tailpipe-plugin-sdk/enrichment"
	"github.com/turbot/tailpipe-plugin-sdk/row_source"
	"github.com/turbot/tailpipe-plugin-sdk/artifact_source"
	"github.com/turbot/tailpipe-plugin-sdk/helpers"
	"github.com/turbot/tailpipe-plugin-sdk/parse"
	"github.com/turbot/tailpipe-plugin-sdk/table"
	"github.com/turbot/tailpipe-plugin-sdk/types"
)

type AuditLogTable struct {
	table.TableImpl[*rows.AuditLog, *AuditLogTableConfig, *config.GcpConnection]
}

func NewAuditLogTable() table.Table {
	return &AuditLogTable{}
}

func (c *AuditLogTable) Init(ctx context.Context, connectionSchemaProvider table.ConnectionSchemaProvider, req *types.CollectRequest) error {
	// call base init
	if err := c.TableImpl.Init(ctx, connectionSchemaProvider, req); err != nil {
		return err
	}

	c.initMapper()
	return nil
}

func (c *AuditLogTable) initMapper() {
	// TODO switch on source
	c.Mapper = mappers.NewAuditLogMapper()
}

func (c *AuditLogTable) Identifier() string {
	return "gcp_audit_log"
}

func (c *AuditLogTable) GetRowSchema() any {
	return rows.AuditLog{}
}

func (c *AuditLogTable) GetConfigSchema() parse.Config {
	return &AuditLogTableConfig{}
}

func (c *AuditLogTable) GetSourceOptions(sourceType string) []row_source.RowSourceOption {
	return []row_source.RowSourceOption{
		artifact_source.WithRowPerLine(),
	}
}

func (c *AuditLogTable) EnrichRow(row *rows.AuditLog, sourceEnrichmentFields *enrichment.CommonFields) (*rows.AuditLog, error) {

	if sourceEnrichmentFields != nil {
		row.CommonFields = *sourceEnrichmentFields
	}

	row.TpID = xid.New().String()
	row.TpTimestamp = helpers.UnixMillis(row.Timestamp.UnixNano() / int64(time.Millisecond))
	row.TpIngestTimestamp = helpers.UnixMillis(time.Now().UnixNano() / int64(time.Millisecond))
	row.TpIndex = "todo" // TODO: #figure out how to get an accountable identifier for the index
	row.TpDate = row.Timestamp.Format("2006-01-02")

	if row.AuthenticationPrincipal != nil {
		row.TpUsernames = append(row.TpUsernames, *row.AuthenticationPrincipal)
	}
	if row.RequestCallerIp != nil {
		row.TpIps = append(row.TpIps, *row.RequestCallerIp)
		row.TpSourceIP = row.RequestCallerIp
	}

	// TODO: #finish payload.Request is a struct which has `Fields` property of map[string]*Value - how to handle? (common keys: @type / name - but this can literally contain anything!)
	// TODO: #finish payload.AuthorizationInfo is an array of structs with Resource (string), Permission (string), and Granted (bool) properties, seems to mostly be a single item but could be more - best way to handle?

	return row, nil
}