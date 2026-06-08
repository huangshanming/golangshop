package storage

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"goshop/internal/config"
)

type MinIOClient struct {
	client *minio.Client
	cfg    config.MinIOConfig
}

func NewMinIOClient(cfg *config.MinIOConfig) (*MinIOClient, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("minio client: %w", err)
	}
	m := &MinIOClient{client: client, cfg: *cfg}
	if err := m.ensureBucket(context.Background()); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *MinIOClient) ensureBucket(ctx context.Context) error {
	exists, err := m.client.BucketExists(ctx, m.cfg.Bucket)
	if err != nil {
		return err
	}
	if !exists {
		if err := m.client.MakeBucket(ctx, m.cfg.Bucket, minio.MakeBucketOptions{}); err != nil {
			return err
		}
	}
	policy := fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {"AWS": ["*"]},
    "Action": ["s3:GetObject"],
    "Resource": ["arn:aws:s3:::%s/*"]
  }]
}`, m.cfg.Bucket)
	return m.client.SetBucketPolicy(ctx, m.cfg.Bucket, policy)
}

func (m *MinIOClient) PutObject(ctx context.Context, objectKey string, reader io.Reader, size int64, contentType string) (string, error) {
	_, err := m.client.PutObject(ctx, m.cfg.Bucket, objectKey, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", err
	}
	return m.PublicURL(objectKey), nil
}

func (m *MinIOClient) PublicURL(objectKey string) string {
	scheme := "http"
	if m.cfg.UseSSL {
		scheme = "https"
	}
	key := strings.TrimPrefix(objectKey, "/")
	return fmt.Sprintf("%s://%s/%s/%s", scheme, m.cfg.Endpoint, m.cfg.Bucket, key)
}
