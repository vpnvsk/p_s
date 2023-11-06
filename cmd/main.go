package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/vpnvsk/p_s"
	handl "github.com/vpnvsk/p_s/pkg/handler"
	"github.com/vpnvsk/p_s/pkg/repository"
	"github.com/vpnvsk/p_s/pkg/service"
)

func main() {

	if err := initConfig(); err != nil {
		logrus.Fatalf("error while reading config files %s", err.Error())
	}
	if err := godotenv.Load(); err != nil {
		logrus.Fatalf("Failed to load config: %s", err.Error())
	}
	db, err := repository.NewPostgresDb(repository.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		Username: os.Getenv("DB_USERNAME"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
		SSLMode:  os.Getenv("DB_SSLMODE"),
	})
	if err != nil {
		logrus.Fatalf("Failed to connect to database:%s", err.Error())
	}
	serviceConfig := service.ServiceConfig{}
	serviceConfig.SetFields(os.Getenv("salt"), os.Getenv("signingKey"), os.Getenv("key"), 2*time.Hour)
	repos := repository.NewRepository(db)
	services := service.NewService(repos, serviceConfig)
	handler := handl.NewHandler(services)
	srv := new(p_s.Server)
	go func() {
		if err := srv.Run(viper.GetString("port"), handler.InitRoutes()); err != nil {
			logrus.Fatalf("error while starting server %s", err.Error())
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit
	if err = srv.ShutDown(context.Background()); err != nil {
		logrus.Errorf("error while shuting down: %s", err.Error())
	}
	db.Close()
}

func initConfig() error {
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}