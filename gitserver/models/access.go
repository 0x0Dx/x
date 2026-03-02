package models

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

const (
	AU_READABLE = iota + 1
	AU_WRITABLE
)

type Access struct {
	ID       int64     `gorm:"primaryKey"`
	UserName string    `gorm:"uniqueIndex:idx_user_repo"`
	RepoName string    `gorm:"uniqueIndex:idx_user_repo"`
	Mode     int       `gorm:"uniqueIndex:idx_user_repo"`
	Created  time.Time `gorm:"autoCreateTime"`
}

func (Access) TableName() string {
	return "access"
}

func AddAccess(access *Access) error {
	return DB.Create(access).Error
}

func HasAccess(userName, repoName string, mode int) (bool, error) {
	var access Access
	err := DB.Where("user_name = ? AND repo_name = ? AND mode = ?",
		strings.ToLower(userName), strings.ToLower(repoName), mode).First(&access).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
