package tables

import (
	"context"
	"time"

	"github.com/rs/xid"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/turbot/tailpipe-plugin-gcp/config"
	"github.com/turbot/tailpipe-plugin-gcp/mappers"
	"github.com/turbot/tailpipe-plugin-gcp/rows"
	"github.com/turbot/tailpipe-plugin-sdk/artifact_source"
	"github.com/turbot/tailpipe-plugin-sdk/artifact_source_config"
	"github.com/turbot/tailpipe-plugin-sdk/enrichment"
	"github.com/turbot/tailpipe-plugin-sdk/helpers"
	"github.com/turbot/tailpipe-plugin-sdk/parse"
	"github.com/turbot/tailpipe-plugin-sdk/row_source"
	"github.com/turbot/tailpipe-plugin-sdk/table"
	"github.com/turbot/tailpipe-plugin-sdk/types"
)

type AuditActivityLogTable struct {
	table.TableImpl[*rows.AuditActivityLog, *AuditActivityLogTableConfig, *config.GcpConnection]
}

func NewAuditActivityLogTable() table.Table {
	return &AuditActivityLogTable{}
}

func (c *AuditActivityLogTable) Init(ctx context.Context, connectionSchemaProvider table.ConnectionSchemaProvider, req *types.CollectRequest) error {
	// call base init
	if err := c.TableImpl.Init(ctx, connectionSchemaProvider, req); err != nil {
		return err
	}

	c.initMapper()
	return nil
}

func (c *AuditActivityLogTable) initMapper() {
	// TODO switch on source
	c.Mapper = mappers.NewAuditActivityLogMapper()
}

func (c *AuditActivityLogTable) Identifier() string {
	return "gcp_audit_activity_log"
}

func (c *AuditActivityLogTable) GetRowSchema() any {
	return rows.AuditActivityLog{}
}

func (c *AuditActivityLogTable) GetConfigSchema() parse.Config {
	return &AuditActivityLogTableConfig{}
}

func (c *AuditActivityLogTable) GetSourceOptions(sourceType string) []row_source.RowSourceOption {
	var opts []row_source.RowSourceOption
	// return []row_source.RowSourceOption{
	// 	artifact_source.WithRowPerLine(),
	// }
	// sourceTypeInfo := artifact_source_config.ArtifactSourceConfig.GetFileLayout()
	// slog.Info("Config info:", sourceTypeInfo)
	switch sourceType {
	case artifact_source.GcpStorageBucketSourceIdentifier:
		defaultArtifactConfig := &artifact_source_config.ArtifactSourceConfigBase{
			FileLayout: utils.ToStringPointer("^cloudaudit\\.googleapis\\.com/activity/\\d{4}/\\d{2}/\\d{2}/\\d{2}:\\d{2}:\\d{2}_\\d{2}:\\d{2}:\\d{2}_S\\d+\\.json$"),
		}
		opts = append(opts, artifact_source.WithDefaultArtifactSourceConfig(defaultArtifactConfig), artifact_source.WithRowPerLine())
	}

	opts = append(opts, artifact_source.WithRowPerLine())

	return opts
}

func (c *AuditActivityLogTable) EnrichRow(row *rows.AuditActivityLog, sourceEnrichmentFields *enrichment.CommonFields) (*rows.AuditActivityLog, error) {

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
