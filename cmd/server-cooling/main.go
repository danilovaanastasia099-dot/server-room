package main

import (
	"fmt"

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

	rep, errRep := repository.New(postgresString)
	if errRep != nil {
		logrus.Fatalf("Error initializing repository: %v", errRep)
	}

	hand := handler.NewHandler(rep)

	application := pkg.NewApp(conf, router, hand)
	application.RunApp()
}
