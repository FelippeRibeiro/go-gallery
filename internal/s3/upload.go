package s3client

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

const BucketName = "go-gallery"

func getEndpoint() string {
	ep := os.Getenv("S3_ENDPOINT")
	if ep == "" {
		return "http://localhost:9000"
	}
	return ep
}

func EnsureBucket(ctx context.Context) error {
	client := ObterS3Client()
	_, err := client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(BucketName)})
	if err == nil {
		return nil
	}
	_, err = client.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(BucketName)})
	return err
}

func UploadPublic(ctx context.Context, key string, body []byte, contentType string) error {
	_, err := ObterS3Client().PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(BucketName),
		Key:         aws.String(key),
		Body:        bytes.NewReader(body),
		ContentType: aws.String(contentType),
		ACL:         types.ObjectCannedACLPublicRead,
	})
	return err
}

func UploadPrivate(ctx context.Context, key string, body []byte, contentType string) error {
	_, err := ObterS3Client().PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(BucketName),
		Key:         aws.String(key),
		Body:        bytes.NewReader(body),
		ContentType: aws.String(contentType),
		ACL:         types.ObjectCannedACLPrivate,
	})
	return err
}

// ObjectURL returns the public URL for a given S3 key.
func ObjectURL(key string) string {
	return fmt.Sprintf("%s/%s/%s", getEndpoint(), BucketName, key)
}

// AlbumPhotoKeys returns the S3 keys for a photo within an album.
// Structure: albums/{albumID}/original-{photoID}.{ext}
//
//	albums/{albumID}/preview-{photoID}.jpg
func AlbumPhotoKeys(albumID int64, photoID, ext string) (originalKey, previewKey string) {
	originalKey = fmt.Sprintf("albums/%d/original-%s.%s", albumID, photoID, ext)
	previewKey = fmt.Sprintf("albums/%d/preview-%s.jpg", albumID, photoID)
	return
}

// DeleteObject removes an object from S3. Errors are silently ignored if the key doesn't exist.
func DeleteObject(ctx context.Context, key string) error {
	_, err := ObterS3Client().DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(BucketName),
		Key:    aws.String(key),
	})
	return err
}

// AlbumCoverKey returns the S3 key for an album's cover image.
func AlbumCoverKey(albumID int64) string {
	return fmt.Sprintf("albums/%d/cover.jpg", albumID)
}

// KeyFromURL extracts the S3 object key from a full URL.
func KeyFromURL(url string) string {
	prefix := getEndpoint() + "/" + BucketName + "/"
	return strings.TrimPrefix(url, prefix)
}

// PresignGetObject generates a pre-signed GET URL for a private S3 object.
func PresignGetObject(ctx context.Context, key string, expires time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(ObterS3Client())
	req, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(BucketName),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expires))
	if err != nil {
		return "", err
	}
	return req.URL, nil
}

// EnsurePublicReadPolicy sets a bucket policy allowing public GET on preview and cover files.
func EnsurePublicReadPolicy(ctx context.Context) error {
	policy := `{
		"Version": "2012-10-17",
		"Statement": [{
			"Effect": "Allow",
			"Principal": {"AWS": ["*"]},
			"Action": ["s3:GetObject"],
			"Resource": [
				"arn:aws:s3:::go-gallery/albums/*/preview-*.jpg",
				"arn:aws:s3:::go-gallery/albums/*/cover.jpg",
				"arn:aws:s3:::go-gallery/profiles/*/avatar.jpg"
			]
		}]
	}`
	_, err := ObterS3Client().PutBucketPolicy(ctx, &s3.PutBucketPolicyInput{
		Bucket: aws.String(BucketName),
		Policy: aws.String(policy),
	})
	return err
}
