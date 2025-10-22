package database

import (
	"context"
	"database/sql"

	"github.com/lva100/go-service/config"
	"github.com/lva100/go-service/pkg/logger"
	_ "github.com/microsoft/go-mssqldb"
)

func CreateDbPool(config *config.DatabaseConfig, logger *logger.Logger) (*sql.DB, error) {
	var err error
	dbpool, err := sql.Open("sqlserver", config.Url)
	if err != nil {
		logger.Error("Не удалось подключиться к базе данных", err)
		panic(err)
	}
	dbpool.SetConnMaxLifetime(0)
	ctx := context.Background()
	err = dbpool.PingContext(ctx)
	if err != nil {
		logger.Error("Не удалось подключиться к базе данных", err)
	}
	logger.Info("Подключились к базе данных")
	return dbpool, nil
}
