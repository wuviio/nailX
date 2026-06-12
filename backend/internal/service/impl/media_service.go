package impl

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/nailx/backend/internal/service"
)

type mediaService struct {
	s3Client   *s3.Client
	bucketName string
	cdnBase    string
}

func NewMediaService(bucketName, cdnBase string) (service.MediaService, error) {
	cfg, err := awsconfig.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, err
	}
	return &mediaService{
		s3Client:   s3.NewFromConfig(cfg),
		bucketName: bucketName,
		cdnBase:    cdnBase,
	}, nil
}

var purposePrefix = map[string]string{
	"ar_snapshot":    "ar/",
	"portfolio":      "portfolio/",
	"review_photo":   "reviews/",
	"design_preview": "designs/",
}

// GeneratePresignedURL は fileType に MIME タイプ（例: "image/jpeg"）を受け取る。
// S3 キーには MIME タイプから導出した拡張子を使用し、ContentType には MIME タイプをそのまま設定する。
func (s *mediaService) GeneratePresignedURL(ctx context.Context, userID, fileType, purpose string) (uploadURL, fileURL string, err error) {
	prefix, ok := purposePrefix[purpose]
	if !ok {
		return "", "", fmt.Errorf("unknown purpose: %s", purpose)
	}

	// "image/jpeg" → "jpeg", "image/png" → "png" のように拡張子を導出
	ext := extFromMIME(fileType)
	key := prefix + userID + "/" + uuid.NewString() + "." + ext

	presignClient := s3.NewPresignClient(s.s3Client)
	presignReq, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(key),
		ContentType: aws.String(fileType), // すでに MIME タイプなのでそのまま使用
	}, s3.WithPresignExpires(15*time.Minute))
	if err != nil {
		return "", "", err
	}

	fileURL = s.cdnBase + "/" + key
	return presignReq.URL, fileURL, nil
}

// extFromMIME は MIME タイプ（例: "image/jpeg"）からファイル拡張子（例: "jpeg"）を返す
func extFromMIME(mimeType string) string {
	parts := strings.SplitN(mimeType, "/", 2)
	if len(parts) != 2 || parts[1] == "" {
		return "bin"
	}
	// image/jpeg → jpeg, image/png → png, image/webp → webp など
	return parts[1]
}
