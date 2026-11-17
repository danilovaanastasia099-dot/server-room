package repository

import (
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type RepositorySettings struct {
	PostgresDSN      string
	MinioEndpoint    string
	MinioAccessKey   string
	MinioSecretKey   string
	MinioImageBucket string
	MinioVideoBucket string
}

type Repository struct {
	db               *gorm.DB
	minio            *minio.Client
	minioEndpoint    string
	minioImageBucket string
	minioVideoBucket string
}

func New(settings *RepositorySettings) (*Repository, error) {
	db, err := gorm.Open(postgres.Open(settings.PostgresDSN), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к PostgreSQL: %w", err)
	}

	useSSL := false
	endpoint := settings.MinioEndpoint
	if endpoint == "" {
		endpoint = "localhost:9000"
	}

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(settings.MinioAccessKey, settings.MinioSecretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		logrus.Warnf("Ошибка инициализации клиента MinIO: %v", err)
	} else {
		for _, b := range []string{settings.MinioImageBucket, settings.MinioVideoBucket} {
			if b != "" {
				exists, errExists := minioClient.BucketExists(context.Background(), b)
				if errExists == nil && !exists {
					_ = minioClient.MakeBucket(context.Background(), b, minio.MakeBucketOptions{})
				}
			}
		}
	}

	repo := &Repository{
		db:               db,
		minio:            minioClient,
		minioEndpoint:    endpoint,
		minioImageBucket: settings.MinioImageBucket,
		minioVideoBucket: settings.MinioVideoBucket,
	}

	return repo, nil
}

func (r *Repository) DB() *gorm.DB {
	return r.db
}

func (r *Repository) MinIO() *minio.Client {
	return r.minio
}

func (r *Repository) UploadFileToMinIO(bucketName, prefix string, userID uint, header *multipart.FileHeader) (string, error) {
	if r.minio == nil {
		return "", fmt.Errorf("minio клиент не инициализирован")
	}

	exists, errExists := r.minio.BucketExists(context.Background(), bucketName)
	if errExists == nil && !exists {
		_ = r.minio.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{})
	}

	file, err := header.Open()
	if err != nil {
		return "", fmt.Errorf("ошибка открытия файла: %w", err)
	}
	defer file.Close()

	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && n == 0 {
		return "", fmt.Errorf("ошибка чтения файла: %w", err)
	}
	contentType := http.DetectContentType(buffer[:n])

	_, err = file.Seek(0, 0)
	if err != nil {
		return "", fmt.Errorf("ошибка сброса указателя чтения файла: %w", err)
	}

	origExt := strings.ToLower(filepath.Ext(header.Filename))
	if origExt == "" || len(origExt) > 5 {
		if strings.Contains(contentType, "jpeg") || strings.Contains(contentType, "jpg") {
			origExt = ".jpg"
		} else if strings.Contains(contentType, "png") {
			origExt = ".png"
		} else if strings.Contains(contentType, "mp4") {
			origExt = ".mp4"
		} else {
			origExt = ".bin"
		}
	}

	latinFilename := fmt.Sprintf("comp_%s_u%d_%d%s", prefix, userID, time.Now().UnixNano(), origExt)

	_, err = r.minio.PutObject(
		context.Background(),
		bucketName,
		latinFilename,
		file,
		header.Size,
		minio.PutObjectOptions{
			ContentType: contentType,
		},
	)
	if err != nil {
		return "", fmt.Errorf("не удалось загрузить объект в MinIO (%s): %w", bucketName, err)
	}

	fileURL := fmt.Sprintf("http://%s/%s/%s", r.minioEndpoint, bucketName, latinFilename)
	logrus.Infof("Файл успешно загружен в MinIO: %s", fileURL)
	return fileURL, nil
}
