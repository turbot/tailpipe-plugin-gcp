package sources

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"

	"github.com/turbot/tailpipe-plugin-sdk/artifact_source_config"
)

// GcpStorageBucketSourceConfig is the configuration for [GcpStorageBucketSource]
type GcpStorageBucketSourceConfig struct {
	artifact_source_config.ArtifactSourceConfigBase
	// required to allow partial decoding
	Remain hcl.Body `hcl:",remain" json:"-"`

	Bucket     string   `hcl:"bucket"`
	Prefix     string   `hcl:"prefix"`
	Extensions []string `hcl:"extensions,optional"`
}

func (g *GcpStorageBucketSourceConfig) Validate() error {
	if g.Bucket == "" {
		return fmt.Errorf("bucket is required and cannot be empty")
	}

	// Check format of extensions
	var invalidExtensions []string
	for _, e := range g.Extensions {
		if len(e) == 0 {
			invalidExtensions = append(invalidExtensions, "<empty>")
		} else if e[0] != '.' {
			invalidExtensions = append(invalidExtensions, e)
		}
	}
	if len(invalidExtensions) > 0 {
		return fmt.Errorf("invalid extensions: %s", strings.Join(invalidExtensions, ","))
	}

	return nil
}

func (*GcpStorageBucketSourceConfig) Identifier() string {
	return GcpStorageBucketSourceIdentifier
}
