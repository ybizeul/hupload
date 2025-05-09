package storage

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"path"
	"sort"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinioStorageConfig is the configuration structure for the s3 backend
// AWSKey and AWSSecret are the credentials to access the bucket
// Bucket is the Bucket name
// MaxFileSize is the maximum size in MB for an item
// MaxShareSize is the maximum size in MB for a share
type MinioStorageConfig struct {
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
type MinioBackend struct {
	Options             MinioStorageConfig
	DefaultValidityDays int

	Client *minio.Client
}

// NewFileStorage creates a new FileBackend with the provided options o
func NewMinioStorage(o MinioStorageConfig) *MinioBackend {
	r := MinioBackend{
		Options: o,
	}

	err := r.initialize()
	if err != nil {
		return nil
	}

	return &r
}

func (b *MinioBackend) initialize() error {
	c, err := minio.New(b.Options.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(b.Options.AWSKey, b.Options.AWSSecret, ""),
		Secure: true,
	})
	if err != nil {
		return err
	}

	b.Client = c

	err = b.Client.MakeBucket(context.Background(), b.Options.Bucket, minio.MakeBucketOptions{
		Region: b.Options.Region})
	if err != nil {
		errResp := minio.ToErrorResponse(err)
		if errResp.Code == "BucketAlreadyOwnedByYou" {
			return nil
		}
		return err
	}

	return nil
}

// Migrate will be called at initialization to give an opportunity to
// the backend to migrate data from a previous version to the current one
func (b *MinioBackend) Migrate() error {
	return nil
}

// CreateShare creates a new share
func (b *MinioBackend) CreateShare(ctx context.Context, name, owner string, options Options) (*Share, error) {
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
	j, err := json.Marshal(share)
	if err != nil {
		return nil, err
	}
	r := bytes.NewReader(j)

	_, err = b.Client.PutObject(ctx, b.Options.Bucket, path, r, int64(len(j)), minio.PutObjectOptions{UserMetadata: map[string]string{
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
func (b *MinioBackend) UpdateShare(ctx context.Context, name string, options *Options) (*Options, error) {
	if !IsShareNameSafe(name) {
		return nil, ErrInvalidShareName
	}

	share, err := b.GetShare(ctx, name)
	if err != nil {
		return nil, err
	}

	share.Options = *options

	path := path.Join("shares", name, ".metadata")
	j, err := json.Marshal(share)
	if err != nil {
		return nil, err
	}
	r := bytes.NewReader(j)

	_, err = b.Client.PutObject(ctx, b.Options.Bucket, path, r, int64(len(j)), minio.PutObjectOptions{UserMetadata: map[string]string{
		"metadata": "true",
		"owner":    share.Owner,
		"name":     name,
	},
	})
	if err != nil {
		return nil, err
	}

	return &share.Options, nil
}

// CreateItem creates a new item in a share
func (b *MinioBackend) CreateItem(ctx context.Context, name, item string, size int64, r io.Reader) (*Item, error) {
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
		src = bufio.NewReader(io.LimitReader(r, maxWrite))
	}

	path := path.Join(name, item)

	_, err = b.Client.PutObject(ctx, b.Options.Bucket, path, src, size, minio.PutObjectOptions{})
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
func (b *MinioBackend) DeleteItem(ctx context.Context, share, item string) error {
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

	err = b.Client.RemoveObject(ctx, b.Options.Bucket, path, minio.RemoveObjectOptions{})
	if err != nil {
		return err
	}

	err = b.updateMetadata(ctx, share)
	if err != nil {
		return err
	}

	return nil
}

// GetShare returns the share identified by share
func (b *MinioBackend) GetShare(ctx context.Context, name string) (*Share, error) {
	if !IsShareNameSafe(name) {
		return nil, ErrInvalidShareName
	}
	path := path.Join("shares", name, ".metadata")
	output, err := b.Client.GetObject(ctx, b.Options.Bucket, path, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}

	result := &Share{}
	err = json.NewDecoder(output).Decode(result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// ListShares returns the list of shares available
func (b *MinioBackend) ListShares(ctx context.Context) ([]Share, error) {
	prefix := "shares/"
	output := b.Client.ListObjects(ctx, b.Options.Bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	result := []Share{}
	for item := range output {
		gOutput, err := b.Client.GetObject(ctx, b.Options.Bucket, item.Key, minio.GetObjectOptions{})
		if err != nil {
			return nil, err
		}
		share := &Share{}
		err = json.NewDecoder(gOutput).Decode(share)
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
func (b *MinioBackend) ListShare(ctx context.Context, name string) ([]Item, error) {
	if !IsShareNameSafe(name) {
		return nil, ErrInvalidShareName
	}
	output := b.Client.ListObjects(ctx, b.Options.Bucket, minio.ListObjectsOptions{
		Prefix:    name + "/",
		Recursive: true,
	})

	result := []Item{}
	for infos := range output {
		// inputs := s3.HeadObjectInput{
		// 	Bucket: &b.Options.Bucket,
		// 	Key:    item.Key,
		// }

		// gOutput, err := b.Client.StatObject(ctx, &inputs)
		// if err != nil {
		// 	return nil, err
		// }

		item := &Item{
			Path: infos.Key,
			ItemInfo: ItemInfo{
				Size:         infos.Size,
				DateModified: infos.LastModified,
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
func (b *MinioBackend) DeleteShare(ctx context.Context, name string) error {
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

	err = b.Client.RemoveObject(ctx, b.Options.Bucket, path, minio.RemoveObjectOptions{})
	if err != nil {
		return err
	}

	return nil
}

// GetItem returns the item identified by share and item
func (b *MinioBackend) GetItem(ctx context.Context, share, item string) (*Item, error) {
	if !IsShareNameSafe(share) {
		return nil, ErrInvalidShareName
	}
	if !isItemNameSafe(item) {
		return nil, ErrInvalidItemName
	}

	path := path.Join(share, item)

	aOutput, err := b.Client.GetObjectAttributes(ctx, b.Options.Bucket, path, minio.ObjectAttributesOptions{})

	if err != nil {
		return nil, err
	}

	result := &Item{
		Path: path,
		ItemInfo: ItemInfo{
			DateModified: aOutput.LastModified,
		},
	}
	result.ItemInfo.Size = int64(aOutput.ObjectSize)

	return result, nil
}

// GetItem returns the item identified by share and item
func (b *MinioBackend) GetItemData(ctx context.Context, share, item string) (io.ReadCloser, error) {
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

	aOutput, err := b.Client.GetObject(ctx, b.Options.Bucket, path, minio.GetObjectOptions{})

	if err != nil {
		return nil, err
	}

	return aOutput, err
}

func (b *MinioBackend) updateMetadata(ctx context.Context, s string) error {
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

	j, err := json.Marshal(share)
	if err != nil {
		return err
	}

	path := path.Join("shares", s, ".metadata")

	r := bytes.NewReader(j)

	_, err = b.Client.PutObject(ctx, b.Options.Bucket, path, r, int64(len(j)), minio.PutObjectOptions{})
	if err != nil {
		return err
	}

	return nil
}
