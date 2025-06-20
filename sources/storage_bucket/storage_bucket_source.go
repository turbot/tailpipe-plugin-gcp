package storage_bucket

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path"

	"cloud.google.com/go/storage"
	"github.com/elastic/go-grok"
	"google.golang.org/api/iterator"

	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/pipe-fittings/v2/filter"
	"github.com/turbot/tailpipe-plugin-gcp/config"
	"github.com/turbot/tailpipe-plugin-sdk/artifact_source"
	"github.com/turbot/tailpipe-plugin-sdk/row_source"
	"github.com/turbot/tailpipe-plugin-sdk/types"
)

const GcpStorageBucketSourceIdentifier = "gcp_storage_bucket"

// GcpStorageBucketSource is a [ArtifactSource] implementation that reads artifacts from a GCP Storage bucket
type GcpStorageBucketSource struct {
	artifact_source.ArtifactSourceImpl[*GcpStorageBucketSourceConfig, *config.GcpConnection]

	client    *storage.Client
	errorList []error
}

func (s *GcpStorageBucketSource) Init(ctx context.Context, params *row_source.RowSourceParams, opts ...row_source.RowSourceOption) error {
	// call base to parse config and apply options
	if err := s.ArtifactSourceImpl.Init(ctx, params, opts...); err != nil {
		return err
	}

	client, err := s.getClient(ctx)
	if err != nil {
		return err
	}
	s.client = client

	s.errorList = []error{}

	slog.Info("Initialized GcpStorageBucketSource", "bucket", s.Config.Bucket, "layout", s.Config.FileLayout)
	return nil
}

func (s *GcpStorageBucketSource) Identifier() string {
	return GcpStorageBucketSourceIdentifier
}

func (s *GcpStorageBucketSource) Close() error {
	_ = os.RemoveAll(s.TempDir)
	return s.client.Close()
}

func (s *GcpStorageBucketSource) DiscoverArtifacts(ctx context.Context) error {
	var prefix string
	layout := typehelpers.SafeString(s.Config.GetFileLayout())
	// if there are any optional segments, we expand them into all possible alternatives
	optionalLayouts := artifact_source.ExpandPatternIntoOptionalAlternatives(layout)

	filterMap := make(map[string]*filter.SqlFilter)

	g := grok.New()
	// add any patterns defined in config
	err := g.AddPatterns(s.Config.GetPatterns())
	if err != nil {
		return fmt.Errorf("error adding grok patterns: %v", err)
	}

	if s.Config.Prefix != nil {
		prefix = *s.Config.Prefix
		var newOptionalLayouts []string
		for _, l := range optionalLayouts {
			newOptionalLayouts = append(newOptionalLayouts, fmt.Sprintf("%s%s", prefix, l))
		}
		// Add support for collecting logs from storage buckets that use a flat structure (i.e., without directory-style prefixes).
		// Currently, if a prefix is specified in the config, it is prepended to the layout pattern.
		// For example, if the prefix is "2025-06-06" and the layout is "%{DATA:path}/%{DATA:endpoint}/%{DATA:folder_path}/%{YEAR:year}-%{MONTHNUM:month}-%{MONTHDAY:day}",
		// the resulting layout becomes "2025-06-06%{DATA:path}/%{DATA:endpoint}/%{DATA:folder_path}/%{YEAR:year}-%{MONTHNUM:month}-%{MONTHDAY:day}",
		// which breaks log collection from buckets using a flat file structure.
		// To address this, we're preserving the existing behavior for directory-style buckets,
		// while adding support for flat buckets as a new, optional configuration path.
		optionalLayouts = append(optionalLayouts, newOptionalLayouts...)
	}

	err = s.walk(ctx, s.Config.Bucket, prefix, optionalLayouts, filterMap, g)
	if err != nil {
		s.errorList = append(s.errorList, fmt.Errorf("error discovering artifacts in GCP storage bucket %s, %w", s.Config.Bucket, err))
	}

	if len(s.errorList) > 0 {
		return errors.Join(s.errorList...)
	}

	return nil
}

func (s *GcpStorageBucketSource) DownloadArtifact(ctx context.Context, info *types.ArtifactInfo) error {
	bucket := s.client.Bucket(s.Config.Bucket)
	obj := bucket.Object(info.Name)

	reader, err := obj.NewReader(ctx)
	if err != nil {
		return fmt.Errorf("failed to get object reader: %s", err.Error())
	}
	defer reader.Close()

	localFilePath := path.Join(s.TempDir, info.Name)
	if err := os.MkdirAll(path.Dir(localFilePath), 0755); err != nil {
		return fmt.Errorf("failed to create directory for file, %w", err)
	}

	outFile, err := os.Create(localFilePath)
	if err != nil {
		return fmt.Errorf("failed to create file, %w", err)
	}
	defer outFile.Close()

	if _, err = io.Copy(outFile, reader); err != nil {
		return fmt.Errorf("failed to write data to file, %w", err)
	}

	downloadInfo := types.NewDownloadedArtifactInfo(info, localFilePath, reader.Attrs.Size)

	return s.OnArtifactDownloaded(ctx, downloadInfo)
}

func (s *GcpStorageBucketSource) getClient(ctx context.Context) (*storage.Client, error) {
	opts, err := s.Connection.GetClientOptions(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed setting GCP Storage client config: %s", err.Error())
	}

	client, err := storage.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCP Storage client: %s", err.Error())
	}

	return client, nil
}

func (s *GcpStorageBucketSource) walk(ctx context.Context, bucket string, prefix string, layouts []string, filterMap map[string]*filter.SqlFilter, g *grok.Grok) error {
	bkt := s.client.Bucket(bucket)
	query := &storage.Query{
		Prefix:    prefix,
		Delimiter: "/", // Treat '/' as directory separator
	}

	// List objects and prefixes
	it := bkt.Objects(ctx, query)
	for {
		objAttrs, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return fmt.Errorf("error getting interating next objects, %w", err)
		}

		// Directories
		if objAttrs.Prefix != "" {
			// Process the directory node
			err = s.WalkNode(ctx, objAttrs.Prefix, "", layouts, true, g, filterMap)
			if err != nil {
				if errors.Is(err, fs.SkipDir) {
					continue
				} else {
					return fmt.Errorf("error walking node, %w", err)
				}
			}
			err = s.walk(ctx, bucket, objAttrs.Prefix, layouts, filterMap, g)
			if err != nil {
				s.errorList = append(s.errorList, err)
			}
		}

		// Files
		if objAttrs.Prefix == "" {
			// Process the file node
			err = s.WalkNode(ctx, objAttrs.Name, "", layouts, false, g, filterMap)
			if err != nil {
				s.errorList = append(s.errorList, fmt.Errorf("error parsing object %s, %w", objAttrs.Name, err))
			}
		}
	}

	return nil
}
