package storage

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"path"
	"time"

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
	AWSKey    string `yaml:"aws_key"`
	AWSSecret string `yaml:"aws_secret"`
	Bucket    string `yaml:"bucket"`

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
		context.TODO(),
		config.WithRegion("us-west-2"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(b.Options.AWSKey, b.Options.AWSSecret, "")),
	)
	if err != nil {
		return err
	}

	b.Client = s3.NewFromConfig(c, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	_, err = b.Client.CreateBucket(context.TODO(), &s3.CreateBucketInput{
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
func (b *S3Backend) CreateShare(name, owner string, options Options) (*Share, error) {
	if !isShareNameSafe(name) {
		return nil, ErrInvalidShareName
	}

	_, err := b.GetShare(name)
	if err == nil {
		return nil, ErrShareAlreadyExists
	}

	share := &Share{
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

	_, err = b.Client.PutObject(context.TODO(), &s3.PutObjectInput{
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
func (b *S3Backend) UpdateShare(name string, options *Options) (*Options, error) {
	if !isShareNameSafe(name) {
		return nil, ErrInvalidShareName
	}

	share, err := b.GetShare(name)
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

	_, err = b.Client.PutObject(context.TODO(), &s3.PutObjectInput{
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
func (b *S3Backend) CreateItem(name, item string, size int64, r *bufio.Reader) (*Item, error) {
	if !isShareNameSafe(name) {
		return nil, ErrInvalidShareName
	}
	if !isItemNameSafe(item) {
		return nil, ErrInvalidItemName
	}

	share, err := b.GetShare(name)
	if err != nil {
		return nil, err
	}

	if size == 0 {
		return nil, ErrEmptyFile
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
	_, err = b.Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:        &b.Options.Bucket,
		Key:           &path,
		Body:          src,
		ContentLength: &size,
	})
	if err != nil {
		return nil, err
	}

	err = b.updateMetadata(name)
	if err != nil {
		return nil, err
	}

	result, err := b.GetItem(name, item)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// CreateItem creates a new item in a share
func (b *S3Backend) DeleteItem(share, item string) error {
	if !isShareNameSafe(share) {
		return ErrInvalidShareName
	}
	if !isItemNameSafe(item) {
		return ErrInvalidItemName
	}

	path := path.Join(share, item)

	_, err := b.Client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: &b.Options.Bucket,
		Key:    &path,
	})
	if err != nil {
		return err
	}

	err = b.updateMetadata(share)
	if err != nil {
		return err
	}

	return nil
}

// GetShare returns the share identified by share
func (b *S3Backend) GetShare(name string) (*Share, error) {
	if !isShareNameSafe(name) {
		return nil, ErrInvalidShareName
	}
	path := path.Join("shares", name, ".metadata")
	output, err := b.Client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: &b.Options.Bucket,
		Key:    &path,
	})
	if err != nil {
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
func (b *S3Backend) ListShares() ([]Share, error) {
	prefix := "shares/"
	output, err := b.Client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: &b.Options.Bucket,
		Prefix: &prefix,
	})
	if err != nil {
		return nil, err
	}

	result := []Share{}
	for _, item := range output.Contents {
		gOutput, err := b.Client.GetObject(context.TODO(), &s3.GetObjectInput{
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
	return result, nil
}

// ListShare returns the list of items in a share
func (b *S3Backend) ListShare(name string) ([]Item, error) {
	if !isShareNameSafe(name) {
		return nil, ErrInvalidShareName
	}
	output, err := b.Client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: &b.Options.Bucket,
		Prefix: &name,
	})
	if err != nil {
		return nil, err
	}

	result := []Item{}
	for _, item := range output.Contents {
		inputs := s3.HeadObjectInput{
			Bucket: &b.Options.Bucket,
			Key:    item.Key,

			// ObjectAttributes: []types.ObjectAttributes{
			// 	types.ObjectAttributesObjectSize,
			// },
		}
		gOutput, err := b.Client.HeadObject(context.TODO(), &inputs)
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
	return result, nil
}

// ListShare returns the list of items in a share
func (b *S3Backend) DeleteShare(name string) error {
	if !isShareNameSafe(name) {
		return ErrInvalidShareName
	}

	_, err := b.GetShare(name)
	if err != nil {
		return err
	}

	content, err := b.ListShare(name)
	if err != nil {
		return err
	}
	for _, item := range content {
		err = b.DeleteItem(name, item.Path)
		if err != nil {
			return err
		}
	}

	path := path.Join("shares", name, ".metadata")

	_, err = b.Client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: &b.Options.Bucket,
		Key:    &path,
	})

	if err != nil {
		return err
	}

	return nil
}

// GetItem returns the item identified by share and item
func (b *S3Backend) GetItem(share, item string) (*Item, error) {
	if !isShareNameSafe(share) {
		return nil, ErrInvalidShareName
	}
	if !isItemNameSafe(item) {
		return nil, ErrInvalidItemName
	}

	path := path.Join(share, item)

	aOutput, err := b.Client.GetObjectAttributes(context.TODO(), &s3.GetObjectAttributesInput{
		Bucket: &b.Options.Bucket,
		Key:    &path,
		ObjectAttributes: []types.ObjectAttributes{
			types.ObjectAttributesObjectSize,
		},
	})

	if err != nil {
		return nil, err
	}

	result := &Item{
		Path: path,
		ItemInfo: ItemInfo{
			Size:         *aOutput.ObjectSize,
			DateModified: *aOutput.LastModified,
		},
	}

	return result, nil
}

// GetItem returns the item identified by share and item
func (b *S3Backend) GetItemData(share, item string) (io.ReadCloser, error) {
	if !isShareNameSafe(share) {
		return nil, ErrInvalidShareName
	}
	if !isItemNameSafe(item) {
		return nil, ErrInvalidItemName
	}

	_, err := b.GetItem(share, item)
	if err != nil {
		return nil, err
	}

	path := path.Join(share, item)

	aOutput, err := b.Client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: &b.Options.Bucket,
		Key:    &path,
	})

	if err != nil {
		return nil, err
	}

	return aOutput.Body, err
}

func (b *S3Backend) updateMetadata(s string) error {
	if !isShareNameSafe(s) {
		return ErrInvalidShareName
	}

	c, err := b.ListShare(s)
	if err != nil {
		return err
	}

	count := 0
	capacity := int64(0)
	for _, i := range c {
		count += 1
		capacity += i.ItemInfo.Size
	}

	share, err := b.GetShare(s)
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

	_, err = b.Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: &b.Options.Bucket,
		Key:    &path,
		Body:   j,
	})

	if err != nil {
		return err
	}

	return nil
}
