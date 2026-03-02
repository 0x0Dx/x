package models

import (
	"fmt"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/0x0Dx/x/gitserver/utils"
)

var (
	DB           *gorm.DB
	RepoRootPath string
)

func InitDB() error {
	cfg := utils.Cfg
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Database.User,
		cfg.Database.Passwd,
		cfg.Database.Host,
		cfg.Database.Name,
	)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	return nil
}

func Migrate() error {
	err := DB.AutoMigrate(
		&User{},
		&PublicKey{},
		&Repo{},
		&Access{},
	)
	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}
	return nil
}

func setEngine() {
	RepoRootPath = "/home/git/repositories"
	if err := os.MkdirAll(RepoRootPath, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create repo root path: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	setEngine()
	if err := InitDB(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize database: %v\n", err)
		os.Exit(1)
	}
	if err := Migrate(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to migrate database: %v\n", err)
		os.Exit(1)
	}
}
