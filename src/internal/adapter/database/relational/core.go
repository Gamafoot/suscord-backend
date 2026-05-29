package relational

import (
	"fmt"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlog "gorm.io/gorm/logger"
)

const (
	DialectPostgres = "postgres"
)

func NewConnect(dbURL, log_level string) (*gorm.DB, error) {
	var logger gormlog.Interface

	switch log_level {
	case "info":
		logger = gormlog.Default.LogMode(gormlog.Info)
	case "warn":
		logger = gormlog.Default.LogMode(gormlog.Warn)
	case "error":
		logger = gormlog.Default.LogMode(gormlog.Error)
	default:
		logger = gormlog.Default.LogMode(gormlog.Silent)
	}

	dialect := DialectFromURL(dbURL)

	var (
		db  *gorm.DB
		err error
	)

	switch dialect {
	case DialectPostgres:
		db, err = gorm.Open(postgres.Open(dbURL), &gorm.Config{
			Logger: logger,
		})
	default:
		return nil, fmt.Errorf("unsupported database dialect for url %q", dbURL)
	}
	if err != nil {
		return nil, err
	}

	return db, nil
}

func DialectFromURL(dbURL string) string {
	url := strings.ToLower(strings.TrimSpace(dbURL))

	switch {
	case strings.HasPrefix(url, "postgres://"), strings.HasPrefix(url, "postgresql://"):
		return DialectPostgres
	case strings.Contains(url, "host=") && strings.Contains(url, "dbname="):
		return DialectPostgres
	default:
		return ""
	}
}
