package main

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"server_components/internal/app/config"
	"server_components/internal/app/dsn"
	"server_components/internal/app/handler"
	"server_components/internal/app/repository"
	"server_components/internal/pkg"
)

func main() {
	_ = godotenv.Load()

	router := gin.Default()
	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("Error loading config: %v", err)
	}

	postgresString := dsn.FromEnv()
	fmt.Println("Connecting to PostgreSQL:", postgresString)

	minioEndpoint := os.Getenv("MINIO_ENDPOINT")
	if minioEndpoint == "" {
		minioEndpoint = "localhost:9000"
	}
	minioAccessKey := os.Getenv("MINIO_ACCESS_KEY")
	if minioAccessKey == "" {
		minioAccessKey = "minioadmin"
	}
	minioSecretKey := os.Getenv("MINIO_SECRET_KEY")
	if minioSecretKey == "" {
		minioSecretKey = "minioadmin"
	}
	minioImageBucket := os.Getenv("MINIO_IMAGE_BUCKET")
	if minioImageBucket == "" {
		minioImageBucket = "server-images"
	}
	minioVideoBucket := os.Getenv("MINIO_VIDEO_BUCKET")
	if minioVideoBucket == "" {
		minioVideoBucket = "server-videos"
	}

	settings := &repository.RepositorySettings{
		PostgresDSN:      postgresString,
		MinioEndpoint:    minioEndpoint,
		MinioAccessKey:   minioAccessKey,
		MinioSecretKey:   minioSecretKey,
		MinioImageBucket: minioImageBucket,
		MinioVideoBucket: minioVideoBucket,
	}

	rep, errRep := repository.New(settings)
	if errRep != nil {
		logrus.Fatalf("Error initializing repository: %v", errRep)
	}

	hand := handler.NewHandler(rep)

	application := pkg.NewApp(conf, router, hand)
	application.RunApp()
}
