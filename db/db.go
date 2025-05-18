package db

import (
	"iffbot/utils"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func getDb() (*gorm.DB, error) {
	settings, err := utils.GetSettings()
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(sqlite.Open(settings.Db), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func Migrate() error {
	db, err := getDb()
	if err != nil {
		return err
	}

	if err := db.AutoMigrate(&Chat{}, &Message{}, &Embedding{}); err != nil {
		return err
	}

	return nil
}
