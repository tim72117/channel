// Package store 封裝持久層(GORM + SQLite)。原型階段用 SQLite,
// 之後可換成 Postgres + pgvector(GORM 換 driver 即可,store 介面不變)。
package store

import (
	"errors"
	"fmt"
	"time"

	"github.com/glebarez/sqlite" // 純 Go SQLite driver,免 CGO
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ErrNotFound 是 store 層統一的「查無資料」錯誤。
var ErrNotFound = errors.New("not found")

type Store struct {
	db *gorm.DB
}

// Open 開啟(或建立)資料庫並用 AutoMigrate 套用 schema。
func Open(path string) (*Store, error) {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	// many2many 的 members 中介表由 GORM 從關聯自動建立。
	if err := db.AutoMigrate(&userRow{}, &channelRow{}, &messageRow{}, &entryRow{}); err != nil {
		return nil, fmt.Errorf("automigrate: %w", err)
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// now 統一回傳 UTC 時間。
func now() time.Time { return time.Now().UTC() }
