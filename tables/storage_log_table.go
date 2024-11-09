package tables

import (
	"context"
	"time"

	"github.com/rs/xid"
	"github.com/turbot/tailpipe-plugin-gcp/config"
	"github.com/turbot/tailpipe-plugin-gcp/mappers"
	"github.com/turbot/tailpipe-plugin-gcp/rows"
	"github.com/turbot/tailpipe-plugin-sdk/artifact_source"
	"github.com/turbot/tailpipe-plugin-sdk/enrichment"
	"github.com/turbot/tailpipe-plugin-sdk/helpers"
	"github.com/turbot/tailpipe-plugin-sdk/parse"
	"github.com/turbot/tailpipe-plugin-sdk/row_source"
	"github.com/turbot/tailpipe-plugin-sdk/table"
	"github.com/turbot/tailpipe-plugin-sdk/types"
)

type StorageLogTable struct {
	table.TableImpl[*rows.StorageLog, *StorageLogTableConfig, *config.GcpConnection]
}

func NewStorageLogTable() table.Table {
	return &StorageLogTable{}
}

func (c *StorageLogTable) Init(ctx context.Context, connectionSchemaProvider table.ConnectionSchemaProvider, req *types.CollectRequest) error {
	if err := c.TableImpl.Init(ctx, connectionSchemaProvider, req); err != nil {
		return err
	}
	c.initMapper()
	return nil
}

func (c *StorageLogTable) initMapper() {
	c.Mapper = mappers.NewStorageLogMapper()
}

func (c *StorageLogTable) Identifier() string {
	return "gcp_storage_log"
}

func (c *StorageLogTable) GetRowSchema() any {
	return rows.StorageLog{}
}

func (c *StorageLogTable) GetConfigSchema() parse.Config {
	return &StorageLogTableConfig{}
}

func (c *StorageLogTable) GetSourceOptions(sourceType string) []row_source.RowSourceOption {
	return []row_source.RowSourceOption{
		artifact_source.WithRowPerLine(),
	}
}

func (c *StorageLogTable) EnrichRow(row *rows.StorageLog, sourceEnrichmentFields *enrichment.CommonFields) (*rows.StorageLog, error) {
	if sourceEnrichmentFields != nil {
		row.CommonFields = *sourceEnrichmentFields
	}

	row.TpID = xid.New().String()
	row.TpTimestamp = helpers.UnixMillis(row.Timestamp.UnixNano() / int64(time.Millisecond))
	row.TpIngestTimestamp = helpers.UnixMillis(time.Now().UnixNano() / int64(time.Millisecond))
	row.TpDate = row.Timestamp.Format("2006-01-02")

	return row, nil
}
