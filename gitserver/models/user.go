package models

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/0x0Dx/x/gitserver/utils"
	"golang.org/x/crypto/scrypt"
	"gorm.io/gorm"
)

const (
	UT_INDIVIDUAL = iota + 1
	UT_ORGANIZATION
)

const (
	LT_PLAIN = iota + 1
	LT_LDAP
)

type User struct {
	Id            int64  `gorm:"primaryKey"`
	LowerName     string `gorm:"uniqueIndex;not null"`
	Name          string `gorm:"uniqueIndex;not null"`
	Email         string `gorm:"uniqueIndex;not null"`
	Passwd        string `gorm:"not null"`
	LoginType     int
	Type          int
	NumFollowers  int
	NumFollowings int
	NumStars      int
	NumRepos      int
	Avatar        string    `gorm:"type:varchar(2048)"`
	Created       time.Time `gorm:"autoCreateTime"`
	Updated       time.Time `gorm:"autoUpdateTime"`
}

type Follow struct {
	Id       int64     `gorm:"primaryKey"`
	UserId   int64     `gorm:"uniqueIndex:idx_user_follow"`
	FollowId int64     `gorm:"uniqueIndex:idx_user_follow"`
	Created  time.Time `gorm:"autoCreateTime"`
}

const (
	OP_CREATE_REPO = iota + 1
	OP_DELETE_REPO
	OP_STAR_REPO
	OP_FOLLOW_REPO
	OP_COMMIT_REPO
	OP_PULL_REQUEST
)

type Action struct {
	Id      int64 `gorm:"primaryKey"`
	UserId  int64
	OpType  int
	RepoId  int64
	Content string
	Created time.Time `gorm:"autoCreateTime"`
}

var (
	ErrUserAlreadyExist = errors.New("user already exist")
	ErrUserNotExist     = errors.New("user does not exist")
)

func (User) TableName() string {
	return "user"
}

func IsUserExist(name string) (bool, error) {
	var count int64
	err := DB.Model(&User{}).Where("lower_name = ?", strings.ToLower(name)).Count(&count).Error
	return count > 0, err
}

func RegisterUser(user *User) error {
	isExist, err := IsUserExist(user.Name)
	if err != nil {
		return err
	} else if isExist {
		return ErrUserAlreadyExist
	}

	user.LowerName = strings.ToLower(user.Name)
	user.Avatar = utils.EncodeMd5(user.Email)
	if err := user.EncodePasswd(); err != nil {
		return err
	}
	return DB.Create(user).Error
}

func UpdateUser(user *User) error {
	return DB.Save(user).Error
}

func DeleteUser(user *User) error {
	return DB.Delete(user).Error
}

func (user *User) EncodePasswd() error {
	newPasswd, err := scrypt.Key([]byte(user.Passwd), []byte("!#@FDEWREWR&*("), 16384, 8, 1, 64)
	if err != nil {
		return err
	}
	user.Passwd = fmt.Sprintf("%x", newPasswd)
	return nil
}

func GetUserByKeyId(keyId int64) (*User, error) {
	var user User
	err := DB.Table("public_key").
		Joins("JOIN user ON user.id = public_key.owner_id").
		Where("public_key.id = ?", keyId).
		First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("not exist key owner")
		}
		return nil, err
	}
	return &user, nil
}

func LoginUserPlain(name, passwd string) (*User, error) {
	user := User{Name: name, Passwd: passwd}
	if err := user.EncodePasswd(); err != nil {
		return nil, err
	}

	var foundUser User
	err := DB.Where("name = ? AND passwd = ?", name, user.Passwd).First(&foundUser).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrUserNotExist
		}
		return nil, err
	}
	return &foundUser, nil
}

func FollowUser(userId int64, followId int64) error {
	tx := DB.Begin()
	if err := tx.Create(&Follow{UserId: userId, FollowId: followId}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Model(&User{}).Where("id = ?", followId).UpdateColumn("num_followers", gorm.Expr("num_followers + ?", 1)).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Model(&User{}).Where("id = ?", userId).UpdateColumn("num_followings", gorm.Expr("num_followings + ?", 1)).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func UnFollowUser(userId int64, unFollowId int64) error {
	tx := DB.Begin()
	if err := tx.Where("user_id = ? AND follow_id = ?", userId, unFollowId).Delete(&Follow{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Model(&User{}).Where("id = ?", unFollowId).UpdateColumn("num_followers", gorm.Expr("num_followers - ?", 1)).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Model(&User{}).Where("id = ?", userId).UpdateColumn("num_followings", gorm.Expr("num_followings - ?", 1)).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}
