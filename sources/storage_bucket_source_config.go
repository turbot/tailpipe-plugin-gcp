package sources

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"

	"github.com/turbot/tailpipe-plugin-sdk/artifact_source_config"
)

// GcpStorageBucketSourceConfig is the configuration for [GcpStorageBucketSource]
type GcpStorageBucketSourceConfig struct {
	artifact_source_config.ArtifactSourceConfigImpl
	// required to allow partial decoding
	Remain hcl.Body `hcl:",remain" json:"-"`

	Bucket string  `hcl:"bucket"`
	Prefix *string `hcl:"prefix,optional"`
}

func (g *GcpStorageBucketSourceConfig) Validate() error {
	if g.Bucket == "" {
		return fmt.Errorf("bucket is required and cannot be empty")
	}

	return nil
}

func (*GcpStorageBucketSourceConfig) Identifier() string {
	return GcpStorageBucketSourceIdentifier
}
