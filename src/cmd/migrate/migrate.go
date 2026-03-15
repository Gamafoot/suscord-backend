package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"suscord/internal/config"
	"suscord/internal/infra/database/relational"
	"suscord/internal/infra/database/relational/model"

	pkgerr "github.com/pkg/errors"
	"gorm.io/gorm"
)

var tables = []interface{}{
	model.User{},
	model.Session{},
	model.Chat{},
	model.ChatMember{},
	model.Message{},
	model.Attachment{},
}

func main() {
	log.Println("start migrating...")

	cfg := config.GetConfig()

	db, err := relational.NewConnect(cfg.Database.URL, cfg.Database.LogLevel)
	if err != nil {
		log.Fatalf("failed to connect to database: %+v", err)
	}

	if err = db.AutoMigrate(tables...); err != nil {
		log.Fatalf("failed to migrate models: %v", err)
	}

	scripts, err := getSqlScripts("trigger")
	if err != nil {
		log.Fatalf("failed get trigger scripts: %+v\n", err)
	}

	err = executeScripts(db, scripts)
	if err != nil {
		log.Fatalf("failed to execute triggers: %+v\n", err)
	}

	log.Println("migrate was successed")
}

func getSqlScripts(folderName string) (map[string]string, error) {
	rootDir := fmt.Sprintf("assets/sql/%s/", folderName)

	scripts := make(map[string]string)

	entries, err := os.ReadDir(rootDir)
	if err != nil {
		if os.IsNotExist(err) {
			return scripts, nil
		}
		return nil, pkgerr.WithStack(err)
	}

	for _, entries := range entries {
		if !strings.HasSuffix(entries.Name(), ".sql") {
			continue
		}

		filepath := rootDir + entries.Name()

		content, err := os.ReadFile(filepath)
		if err != nil {
			return nil, pkgerr.WithStack(err)
		}

		scripts[entries.Name()] = string(content)
	}

	return scripts, nil
}

func executeScripts(db *gorm.DB, scripts map[string]string) error {
	tx := db.Begin()

	for filename, script := range scripts {
		if err := tx.Exec(script).Error; err != nil {
			tx.Rollback()
			return pkgerr.Errorf("%s: %v", filename, err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return pkgerr.WithStack(err)
	}

	return nil
}
