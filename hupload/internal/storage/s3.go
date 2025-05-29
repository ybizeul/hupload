package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"path"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// S3StorageConfig is the configuration structure for the s3 backend
// AWSKey and AWSSecret are the credentials to access the bucket
// Bucket is the Bucket name
// MaxFileSize is the maximum size in MB for an item
// MaxShareSize is the maximum size in MB for a share
type S3StorageConfig struct {
	Endpoint     string `yaml:"endpoint,omitempty"`
	UsePathStyle bool   `yaml:"use_path_style,omitempty"`
	AWSKey       string `yaml:"aws_key"`
	AWSSecret    string `yaml:"aws_secret"`
	Bucket       string `yaml:"bucket"`
	Region       string `yaml:"region"`

	MaxFileSize  int64 `yaml:"max_file_mb"`
	MaxShareSize int64 `yaml:"max_share_mb"`
}

// FileBackend is a backend that stores files on the filesystem
// Options is the configuration for the file storage backend
// DefaultValidityDays is a global option in the configuration file that
// sets the default validity of a share in days
type S3Backend struct {
	Options             S3StorageConfig
	DefaultValidityDays int

	Client *s3.Client
}

// NewFileStorage creates a new FileBackend with the provided options o
func NewS3Storage(o S3StorageConfig) *S3Backend {
	r := S3Backend{
		Options: o,
	}

	err := r.initialize()
	if err != nil {
		return nil
	}

	return &r
}

func (b *S3Backend) initialize() error {
	c, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithRegion(b.Options.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(b.Options.AWSKey, b.Options.AWSSecret, "")),
		config.WithHTTPClient(&http.Client{
			Timeout: 0,
		}),
		config.WithRequestChecksumCalculation(aws.RequestChecksumCalculationWhenRequired),
	)
	if err != nil {
		return err
	}

	b.Client = s3.NewFromConfig(c, func(o *s3.Options) {
		if b.Options.UsePathStyle {
			o.UsePathStyle = true
		}
		if b.Options.Endpoint != "" {
			o.BaseEndpoint = &b.Options.Endpoint
		}
	})

	_, err = b.Client.CreateBucket(context.Background(), &s3.CreateBucketInput{
		Bucket: &b.Options.Bucket,
	})
	if err != nil {
		var bne *types.BucketAlreadyOwnedByYou
		if errors.As(err, &bne) {
			return nil
		}
	}

	return nil
}

// Migrate will be called at initialization to give an opportunity to
// the backend to migrate data from a previous version to the current one
func (b *S3Backend) Migrate() error {
	return nil
}

// CreateShare creates a new share
func (b *S3Backend) CreateShare(ctx context.Context, name, owner string, options Options) (*Share, error) {
	if !IsShareNameSafe(name) {
		return nil, ErrInvalidShareName
	}

	_, err := b.GetShare(ctx, name)
	if err == nil {
		return nil, ErrShareAlreadyExists
	}

	if options.Exposure == "" {
		options.Exposure = "upload"
	}

	share := &Share{
		Version:     1,
		Name:        name,
		Owner:       owner,
		Options:     options,
		DateCreated: time.Now(),
	}
	path := path.Join("shares", name, ".metadata")
	j := bytes.NewBuffer([]byte{})
	err = json.NewEncoder(j).Encode(share)
	if err != nil {
		return nil, err
	}

	_, err = b.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &b.Options.Bucket,
		Key:    &path,
		Body:   j,
		Metadata: map[string]string{
			"metadata": "true",
			"owner":    owner,
			"name":     name,
		},
	})

	if err != nil {
		return nil, err
	}

	return share, nil
}

// UpdateShare updates an existing share
func (b *S3Backend) UpdateShare(ctx context.Context, name string, options *Options) (*Options, error) {
	if !IsShareNameSafe(name) {
		return nil, ErrInvalidShareName
	}

	share, err := b.GetShare(ctx, name)
	if err != nil {
		return nil, err
	}

	share.Options = *options

	path := path.Join("shares", name, ".metadata")
	j := bytes.NewBuffer([]byte{})
	err = json.NewEncoder(j).Encode(share)
	if err != nil {
		return nil, err
	}

	_, err = b.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &b.Options.Bucket,
		Key:    &path,
		Body:   j,
	})

	if err != nil {
		return nil, err
	}

	return &share.Options, nil
}

// CreateItem creates a new item in a share
func (b *S3Backend) CreateItem(ctx context.Context, name, item string, size int64, r io.Reader) (*Item, error) {
	if !IsShareNameSafe(name) {
		return nil, ErrInvalidShareName
	}
	if !isItemNameSafe(item) {
		return nil, ErrInvalidItemName
	}

	share, err := b.GetShare(ctx, name)
	if err != nil {
		return nil, err
	}

	// Check amount of free capacity in share according to current limits
	maxWrite := int64(0)

	maxShare := b.Options.MaxShareSize * 1024 * 1024
	if maxShare > 0 {
		maxWrite = maxShare - share.Size
		if maxWrite <= 0 {
			return nil, ErrMaxShareSizeReached
		}
	}

	maxItem := b.Options.MaxFileSize * 1024 * 1024
	if maxItem > 0 {
		if maxItem < size {
			return nil, ErrMaxFileSizeReached
		}
		if maxWrite > maxItem || maxWrite == 0 {
			maxWrite = maxItem
		}
	}

	// maxWrite is the actual allowed size for the item, so we fix the limit
	// to one more byte
	if maxWrite > 0 {
		maxWrite++

		if maxWrite < size {
			return nil, ErrMaxShareSizeReached
		}
	}
	src := r

	// Substitute bufio.Reader with a limited reader
	if maxWrite != 0 {
		src = io.LimitReader(r, maxWrite)
	}

	path := path.Join(name, item)
	input := &s3.PutObjectInput{
		Bucket:        &b.Options.Bucket,
		Key:           &path,
		Body:          src,
		ContentLength: &size,
	}
	_, err = b.Client.PutObject(ctx, input) // s3.WithAPIOptions(
	// 	v4.AddUnsignedPayloadMiddleware,
	// 	v4.RemoveComputePayloadSHA256Middleware,
	// ),

	if err != nil {
		return nil, err
	}

	err = b.updateMetadata(ctx, name)
	if err != nil {
		return nil, err
	}

	result, err := b.GetItem(ctx, name, item)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// CreateItem creates a new item in a share
func (b *S3Backend) DeleteItem(ctx context.Context, share, item string) error {
	if !IsShareNameSafe(share) {
		return ErrInvalidShareName
	}
	if !isItemNameSafe(item) {
		return ErrInvalidItemName
	}

	path := path.Join(share, item)

	_, err := b.GetItem(ctx, share, item)
	if err != nil {
		return err
	}

	_, err = b.Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &b.Options.Bucket,
		Key:    &path,
	})

	if err != nil {
		var bne *types.NoSuchKey
		if errors.As(err, &bne) {
			return ErrItemNotFound
		}
		return err
	}

	err = b.updateMetadata(ctx, share)
	if err != nil {
		return err
	}

	return nil
}

// GetShare returns the share identified by share
func (b *S3Backend) GetShare(ctx context.Context, name string) (*Share, error) {
	if !IsShareNameSafe(name) {
		return nil, ErrInvalidShareName
	}
	path := path.Join("shares", name, ".metadata")
	output, err := b.Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &b.Options.Bucket,
		Key:    &path,
	})
	if err != nil {
		var bne *types.NoSuchKey
		if errors.As(err, &bne) {
			return nil, ErrShareNotFound
		}
		return nil, err
	}

	result := &Share{}
	err = json.NewDecoder(output.Body).Decode(result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// ListShares returns the list of shares available
func (b *S3Backend) ListShares(ctx context.Context) ([]Share, error) {
	prefix := "shares/"
	output, err := b.Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: &b.Options.Bucket,
		Prefix: &prefix,
	})
	if err != nil {
		return nil, err
	}

	result := []Share{}
	for _, item := range output.Contents {
		gOutput, err := b.Client.GetObject(ctx, &s3.GetObjectInput{
			Bucket: &b.Options.Bucket,
			Key:    item.Key,
		})
		if err != nil {
			return nil, err
		}
		share := &Share{}
		err = json.NewDecoder(gOutput.Body).Decode(share)
		if err != nil {
			return nil, err
		}
		result = append(result, *share)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].DateCreated.After(result[j].DateCreated)
	})

	return result, nil
}

// ListShare returns the list of items in a share
func (b *S3Backend) ListShare(ctx context.Context, name string) ([]Item, error) {
	if !IsShareNameSafe(name) {
		return nil, ErrInvalidShareName
	}
	output, err := b.Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: &b.Options.Bucket,
		Prefix: &name,
	})
	if err != nil {
		var bne *types.NoSuchKey
		if errors.As(err, &bne) {
			return nil, ErrShareNotFound
		}
		return nil, err
	}

	result := []Item{}
	for _, item := range output.Contents {
		inputs := s3.HeadObjectInput{
			Bucket: &b.Options.Bucket,
			Key:    item.Key,
		}

		gOutput, err := b.Client.HeadObject(ctx, &inputs)
		if err != nil {
			return nil, err
		}

		item := &Item{
			Path: *item.Key,
			ItemInfo: ItemInfo{
				Size:         *gOutput.ContentLength,
				DateModified: *gOutput.LastModified,
			},
		}
		result = append(result, *item)
	}

	// Sort items by modification date, newest first
	sort.Slice(result, func(i, j int) bool {
		return result[i].ItemInfo.DateModified.After(result[j].ItemInfo.DateModified)
	})

	return result, nil
}

// ListShare returns the list of items in a share
func (b *S3Backend) DeleteShare(ctx context.Context, name string) error {
	if !IsShareNameSafe(name) {
		return ErrInvalidShareName
	}

	_, err := b.GetShare(ctx, name)
	if err != nil {
		return err
	}

	content, err := b.ListShare(ctx, name)
	if err != nil {
		return err
	}
	for _, item := range content {
		err = b.DeleteItem(ctx, name, path.Base(item.Path))
		if err != nil {
			return err
		}
	}

	path := path.Join("shares", name, ".metadata")

	_, err = b.Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &b.Options.Bucket,
		Key:    &path,
	})

	if err != nil {
		return err
	}

	return nil
}

// GetItem returns the item identified by share and item
func (b *S3Backend) GetItem(ctx context.Context, share, item string) (*Item, error) {
	if !IsShareNameSafe(share) {
		return nil, ErrInvalidShareName
	}
	if !isItemNameSafe(item) {
		return nil, ErrInvalidItemName
	}

	path := path.Join(share, item)

	aOutput, err := b.Client.GetObjectAttributes(ctx, &s3.GetObjectAttributesInput{
		Bucket: &b.Options.Bucket,
		Key:    &path,
		ObjectAttributes: []types.ObjectAttributes{
			types.ObjectAttributesObjectSize,
		},
	})

	if err != nil {
		var bne *types.NoSuchKey
		if errors.As(err, &bne) {
			return nil, ErrItemNotFound
		}
		return nil, err
	}

	result := &Item{
		Path: path,
		ItemInfo: ItemInfo{
			DateModified: *aOutput.LastModified,
		},
	}

	if aOutput.ObjectSize != nil {
		result.ItemInfo.Size = *aOutput.ObjectSize
	}

	return result, nil
}

// GetItem returns the item identified by share and item
func (b *S3Backend) GetItemData(ctx context.Context, share, item string) (io.ReadCloser, error) {
	if !IsShareNameSafe(share) {
		return nil, ErrInvalidShareName
	}
	if !isItemNameSafe(item) {
		return nil, ErrInvalidItemName
	}

	_, err := b.GetItem(ctx, share, item)
	if err != nil {
		return nil, err
	}

	path := path.Join(share, item)

	aOutput, err := b.Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &b.Options.Bucket,
		Key:    &path,
	})

	if err != nil {
		return nil, err
	}

	return aOutput.Body, err
}

func (b *S3Backend) updateMetadata(ctx context.Context, s string) error {
	if !IsShareNameSafe(s) {
		return ErrInvalidShareName
	}

	c, err := b.ListShare(ctx, s)
	if err != nil {
		return err
	}

	count := 0
	capacity := int64(0)
	for _, i := range c {
		count += 1
		capacity += i.ItemInfo.Size
	}

	share, err := b.GetShare(ctx, s)
	if err != nil {
		return err
	}
	share.Count = int64(count)
	share.Size = capacity

	j := bytes.NewBuffer([]byte{})
	err = json.NewEncoder(j).Encode(share)
	if err != nil {
		return err
	}

	path := path.Join("shares", s, ".metadata")

	_, err = b.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &b.Options.Bucket,
		Key:    &path,
		Body:   j,
	})

	if err != nil {
		return err
	}

	return nil
}
