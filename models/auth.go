package models

import (
	"errors"
	"ops-message-unified-push/pkg/util"

	"gorm.io/gorm"
)

type Auth struct {
	ID             int    `json:"id" gorm:"autoIncrement;type:integer;primaryKey"`
	Username       string `json:"username" gorm:"type:varchar(100);default:''"`
	Password       string `json:"password" gorm:"type:varchar(100);default:''"`
	Channel        string `json:"channel" gorm:"type:varchar(20);default:'local'"` // local/oidc
	ExternalSub    string `json:"external_sub" gorm:"type:varchar(255);default:''"`
	ExternalIssuer string `json:"external_issuer" gorm:"type:varchar(255);default:''"`
	ExternalEmail  string `json:"external_email" gorm:"type:varchar(255);default:''"`
}

// CheckAuth 检查用户信息（密码使用 MD5 加密）
func CheckAuth(username, password string) (bool, error) {
	var auth Auth
	err := db.Where("username = ?", username).First(&auth).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	// OIDC 用户不允许本地密码登录
	if auth.Channel == "oidc" {
		return false, errors.New("OIDC_USER_USE_OIDC_LOGIN")
	}

	// 对输入的密码进行 MD5 加密后比较
	encryptedPwd := util.EncodeMD5(password)
	if auth.Password != encryptedPwd {
		return false, nil
	}

	return true, nil
}

func GetUserByUsername(username string) (*Auth, error) {
	var auth Auth
	err := db.Where("username = ?", username).First(&auth).Error
	if err != nil {
		return nil, err
	}
	return &auth, nil
}

func GetUserByID(id int) (*Auth, error) {
	var auth Auth
	err := db.Where("id = ?", id).First(&auth).Error
	if err != nil {
		return nil, err
	}
	return &auth, nil
}

func GetUserByUsernameExcludeID(username string, excludeID int) (*Auth, error) {
	var auth Auth
	query := db.Where("username = ?", username)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}
	err := query.First(&auth).Error
	if err != nil {
		return nil, err
	}
	return &auth, nil
}

func IsUsernameExists(username string, excludeID int) (bool, error) {
	_, err := GetUserByUsernameExcludeID(username, excludeID)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return false, err
}

// EditUser 编辑用户信息
func EditUser(username string, data map[string]interface{}) error {
	if err := db.Model(&Auth{}).Where("username = ? ", username).Updates(data).Error; err != nil {
		return err
	}
	return nil
}

// AddUser 添加本地用户
func AddUser(account string, password string) error {
	// 对密码进行 MD5 加密存储
	auth := Auth{
		Username: account,
		Password: util.EncodeMD5(password),
		Channel:  "local",
	}
	if err := db.Create(&auth).Error; err != nil {
		return err
	}
	return nil
}

// AddCasdoorUser 添加 Casdoor 用户
func AddCasdoorUser(username, externalSub string) (*Auth, error) {
	auth := Auth{
		Username:    username,
		Password:    "",
		Channel:     "casdoor",
		ExternalSub: externalSub,
	}
	if err := db.Create(&auth).Error; err != nil {
		return nil, err
	}
	return &auth, nil
}

// UpdateAuthChannelInfo 更新用户渠道信息
func UpdateAuthChannelInfo(userID int, channel, externalSub string) error {
	return db.Model(&Auth{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"channel":      channel,
		"external_sub": externalSub,
	}).Error
}

// GetUserByChannelAndExternalID 根据渠道和外部ID获取用户
func GetUserByChannelAndExternalID(channel, sub, issuer string) (*Auth, error) {
	var auth Auth
	err := db.Where("channel = ? AND external_sub = ? AND external_issuer = ?", channel, sub, issuer).First(&auth).Error
	if err != nil {
		return nil, err
	}
	return &auth, nil
}

// GetUsersByChannel 根据渠道获取用户列表
func GetUsersByChannel(channel string, pageNum, pageSize int) ([]Auth, error) {
	var auths []Auth
	offset := (pageNum - 1) * pageSize
	err := db.Where("channel = ?", channel).Offset(offset).Limit(pageSize).Find(&auths).Error
	if err != nil {
		return nil, err
	}
	return auths, nil
}

// GetUserCountByChannel 获取指定渠道的用户数量
func GetUserCountByChannel(channel string) (int64, error) {
	var count int64
	err := db.Model(&Auth{}).Where("channel = ?", channel).Count(&count).Error
	return count, err
}

func EditUserByID(id int, data map[string]interface{}) error {
	return db.Model(&Auth{}).Where("id = ?", id).Updates(data).Error
}

func DeleteUserByID(id int) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", id).Delete(&RbacUserRole{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", id).Delete(&RbacUserGroupMember{}).Error; err != nil {
			return err
		}
		// TODO: 后续移除 AuthIdentity 相关逻辑
		if err := tx.Where("user_id = ?", id).Delete(&AuthIdentity{}).Error; err != nil {
			return err
		}
		return tx.Where("id = ?", id).Delete(&Auth{}).Error
	})
}
