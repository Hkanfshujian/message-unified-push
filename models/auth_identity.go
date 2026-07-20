package models

import (
	"errors"
	"strings"

	"gorm.io/gorm"
)

type AuthIdentity struct {
	IDModel
	SoftDeleteModel
	UserID        int    `json:"user_id" gorm:"type:integer;not null;index"`
	Provider      string `json:"provider" gorm:"type:varchar(50);not null;index:idx_provider_sub,unique"`
	ExternalSub   string `json:"external_sub" gorm:"type:varchar(255);not null;index:idx_provider_sub,unique"`
	ExternalEmail string `json:"external_email" gorm:"type:varchar(255);default:'';index"`
	ExternalName  string `json:"external_name" gorm:"type:varchar(255);default:''"`
}

func GetAuthIdentityByProviderSub(provider string, sub string) (*AuthIdentity, error) {
	var identity AuthIdentity
	if err := db.Where("provider = ? AND external_sub = ?", provider, sub).First(&identity).Error; err != nil {
		return nil, err
	}
	return &identity, nil
}

func AddAuthIdentity(identity *AuthIdentity) error {
	var existing AuthIdentity
	err := db.Unscoped().Where("provider = ? AND external_sub = ?", identity.Provider, identity.ExternalSub).First(&existing).Error
	if err == nil {
		return db.Unscoped().Model(&existing).Updates(map[string]interface{}{
			"user_id":        identity.UserID,
			"external_email": identity.ExternalEmail,
			"external_name":  identity.ExternalName,
			"deleted_at":     nil,
		}).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return db.Create(identity).Error
}

func EditAuthIdentityByID(id uint, data map[string]interface{}) error {
	return db.Model(&AuthIdentity{}).Where("id = ?", id).Updates(data).Error
}

type AuthIdentityWithUser struct {
	AuthIdentity
	Username string `json:"username"`
}

func GetAuthIdentityByID(id uint) (*AuthIdentity, error) {
	var identity AuthIdentity
	if err := db.Where("id = ?", id).First(&identity).Error; err != nil {
		return nil, err
	}
	return &identity, nil
}

func DeleteAuthIdentityByID(id uint) error {
	return db.Unscoped().Where("id = ?", id).Delete(&AuthIdentity{}).Error
}

func GetAuthIdentityTotal(provider string, text string) (int64, error) {
	var total int64
	query := db.Model(&AuthIdentity{})
	if strings.TrimSpace(provider) != "" {
		query = query.Where("provider = ?", strings.TrimSpace(provider))
	}
	if strings.TrimSpace(text) != "" {
		like := "%" + strings.TrimSpace(text) + "%"
		query = query.Where("external_sub LIKE ? OR external_email LIKE ? OR external_name LIKE ?", like, like, like)
	}
	if err := query.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func GetAuthIdentities(pageNum int, pageSize int, provider string, text string) ([]AuthIdentityWithUser, error) {
	result := make([]AuthIdentityWithUser, 0)
	query := db.Table(GetSchema(AuthIdentity{}) + " AS ai").
		Select("ai.*, u.username AS username").
		Joins("LEFT JOIN " + GetSchema(Auth{}) + " AS u ON ai.user_id = u.id").
		Where("ai.deleted_at IS NULL").
		Where("u.deleted_at IS NULL").
		Order("ai.id DESC")
	if strings.TrimSpace(provider) != "" {
		query = query.Where("ai.provider = ?", strings.TrimSpace(provider))
	}
	if strings.TrimSpace(text) != "" {
		like := "%" + strings.TrimSpace(text) + "%"
		query = query.Where("ai.external_sub LIKE ? OR ai.external_email LIKE ? OR ai.external_name LIKE ? OR u.username LIKE ?", like, like, like, like)
	}
	if pageSize > 0 {
		query = query.Offset(pageNum).Limit(pageSize)
	}
	if err := query.Scan(&result).Error; err != nil {
		return nil, err
	}
	return result, nil
}
