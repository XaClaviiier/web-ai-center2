package media

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
)

// S3Config adalah konfigurasi koneksi ke RustFS.
type S3Config struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
}

// S3Store menyimpan gambar di bucket RustFS lewat protokol S3 (mode path-style).
type S3Store struct {
	c      *s3.Client
	bucket string
}

var (
	_ Store     = (*S3Store)(nil)
	_ Pendaftar = (*S3Store)(nil)
)

// NewS3 membuat klien. Region ditetapkan us-east-1 (RustFS tidak memakainya, tetapi SDK mewajibkan nilai).
// Checksum tambahan SDK baru hanya dihitung bila diwajibkan, supaya kompatibel dengan server S3 non-AWS.
func NewS3(cfg S3Config) *S3Store {
	awsCfg := aws.Config{
		Region:                     "us-east-1",
		Credentials:                credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
		RequestChecksumCalculation: aws.RequestChecksumCalculationWhenRequired,
		ResponseChecksumValidation: aws.ResponseChecksumValidationWhenRequired,
		HTTPClient:                 &http.Client{Timeout: 30 * time.Second},
	}
	c := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.Endpoint)
		o.UsePathStyle = true
	})
	return &S3Store{c: c, bucket: cfg.Bucket}
}

// PastikanBucket memeriksa bucket. Bila belum ada: dibuat jika buat=true, selain itu galat.
func (s *S3Store) PastikanBucket(ctx context.Context, buat bool) error {
	_, err := s.c.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(s.bucket)})
	if err == nil {
		return nil
	}
	if !adalahTidakAda(err) {
		return fmt.Errorf("memeriksa bucket %q: %w", s.bucket, err)
	}
	if !buat {
		return fmt.Errorf("bucket %q belum ada; buat dulu di RustFS", s.bucket)
	}
	if _, err := s.c.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(s.bucket)}); err != nil {
		return fmt.Errorf("membuat bucket %q: %w", s.bucket, err)
	}
	return nil
}

func (s *S3Store) Put(ctx context.Context, key string, data []byte, contentType string) error {
	_, err := s.c.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(key),
		Body:          bytes.NewReader(data),
		ContentLength: aws.Int64(int64(len(data))),
		ContentType:   aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("PutObject %s: %w", key, err)
	}
	return nil
}

func (s *S3Store) Get(ctx context.Context, key string) (Objek, error) {
	out, err := s.c.GetObject(ctx, &s3.GetObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(key)})
	if err != nil {
		if adalahTidakAda(err) {
			return Objek{}, ErrTidakAda
		}
		return Objek{}, fmt.Errorf("GetObject %s: %w", key, err)
	}
	return Objek{
		Meta: Meta{ContentType: aws.ToString(out.ContentType), Ukuran: aws.ToInt64(out.ContentLength), ETag: aws.ToString(out.ETag)},
		Body: out.Body,
	}, nil
}

func (s *S3Store) Head(ctx context.Context, key string) (Meta, error) {
	out, err := s.c.HeadObject(ctx, &s3.HeadObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(key)})
	if err != nil {
		if adalahTidakAda(err) {
			return Meta{}, ErrTidakAda
		}
		return Meta{}, fmt.Errorf("HeadObject %s: %w", key, err)
	}
	return Meta{ContentType: aws.ToString(out.ContentType), Ukuran: aws.ToInt64(out.ContentLength), ETag: aws.ToString(out.ETag)}, nil
}

func (s *S3Store) Hapus(ctx context.Context, key string) error {
	if _, err := s.c.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(key)}); err != nil {
		return fmt.Errorf("DeleteObject %s: %w", key, err)
	}
	return nil
}

func (s *S3Store) Daftar(ctx context.Context, prefix string) ([]Info, error) {
	var out []Info
	p := s3.NewListObjectsV2Paginator(s.c, &s3.ListObjectsV2Input{Bucket: aws.String(s.bucket), Prefix: aws.String(prefix)})
	for p.HasMorePages() {
		page, err := p.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("ListObjectsV2 %q: %w", prefix, err)
		}
		for _, o := range page.Contents {
			out = append(out, Info{Key: aws.ToString(o.Key), Ukuran: aws.ToInt64(o.Size), Diubah: aws.ToTime(o.LastModified)})
		}
	}
	return out, nil
}

// adalahTidakAda mengenali galat "objek/bucket tidak ada" dari berbagai bentuk respons S3.
func adalahTidakAda(err error) bool {
	var nsk *types.NoSuchKey
	var nf *types.NotFound
	var nsb *types.NoSuchBucket
	if errors.As(err, &nsk) || errors.As(err, &nf) || errors.As(err, &nsb) {
		return true
	}
	var api smithy.APIError
	if errors.As(err, &api) {
		switch api.ErrorCode() {
		case "NoSuchKey", "NotFound", "NoSuchBucket", "404":
			return true
		}
	}
	return false
}
