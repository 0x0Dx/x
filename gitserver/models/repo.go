package models

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"
)

type Repo struct {
	ID        int64 `gorm:"primaryKey"`
	OwnerId   int64 `gorm:"index"`
	ForkId    int64
	LowerName string `gorm:"uniqueIndex;not null"`
	Name      string `gorm:"index;not null"`
	NumWatchs int
	NumStars  int
	NumForks  int
	Created   time.Time `gorm:"autoCreateTime"`
	Updated   time.Time `gorm:"autoUpdateTime"`
}

func (Repo) TableName() string {
	return "repo"
}

func IsRepositoryExist(user *User, repoName string) (bool, error) {
	var repo Repo
	err := DB.Where("owner_id = ? AND lower_name = ?", user.Id, strings.ToLower(repoName)).First(&repo).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			s, err := os.Stat(filepath.Join(RepoRootPath, user.Name, repoName))
			if err != nil {
				return false, nil
			}
			return s.IsDir(), nil
		}
		return false, err
	}
	s, err := os.Stat(filepath.Join(RepoRootPath, user.Name, repoName))
	if err != nil {
		return false, err
	}
	return s.IsDir(), nil
}

func CreateRepository(user *User, repoName string) (*Repo, error) {
	p := filepath.Join(RepoRootPath, user.Name)
	if err := os.MkdirAll(p, os.ModePerm); err != nil {
		return nil, err
	}
	f := filepath.Join(p, repoName+".git")
	cmd := exec.Command("git", "init", "--bare", f)
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	repo := Repo{OwnerId: user.Id, Name: repoName, LowerName: strings.ToLower(repoName)}
	tx := DB.Begin()
	if err := tx.Create(&repo).Error; err != nil {
		os.RemoveAll(f)
		tx.Rollback()
		return nil, err
	}
	if err := tx.Model(&User{}).Where("id = ?", user.Id).UpdateColumn("num_repos", gorm.Expr("num_repos + ?", 1)).Error; err != nil {
		os.RemoveAll(f)
		tx.Rollback()
		return nil, err
	}
	if err := tx.Commit().Error; err != nil {
		os.RemoveAll(f)
		tx.Rollback()
		return nil, err
	}
	return &repo, nil
}

func GetRepositories(user *User) ([]Repo, error) {
	var repos []Repo
	err := DB.Where("owner_id = ?", user.Id).Find(&repos).Error
	return repos, err
}

func StarRepository(user *User, repoName string) error {
	return nil
}

func UnStarRepository() {}

func WatchRepository() {}

func UnWatchRepository() {}

func DeleteRepository(user *User, repoName string) error {
	tx := DB.Begin()
	if err := tx.Where("owner_id = ? AND name = ?", user.Id, repoName).Delete(&Repo{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Model(&User{}).Where("id = ?", user.Id).UpdateColumn("num_repos", gorm.Expr("num_repos - ?", 1)).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := os.RemoveAll(filepath.Join(RepoRootPath, user.Name, repoName+".git")); err != nil {
		return err
	}
	return nil
}
