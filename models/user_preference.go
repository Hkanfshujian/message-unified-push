package models

import (
	"errors"

	"gorm.io/gorm"
)

type UserPreference struct {
	IDModel
	UserID     int    `json:"user_id" gorm:"type:int;not null;uniqueIndex"`
	ThemeColor string `json:"theme_color" gorm:"type:varchar(30);default:''"`
	ThemeMode  string `json:"theme_mode" gorm:"type:varchar(20);default:'system'"`
	SidebarBg  string `json:"sidebar_bg" gorm:"type:varchar(30);default:'#0b3c51'"`
}

func GetUserPreferenceByUserID(userID int) (*UserPreference, error) {
	var pref UserPreference
	if err := db.Where("user_id = ?", userID).First(&pref).Error; err != nil {
		return nil, err
	}
	return &pref, nil
}

func UpsertUserPreference(userID int, themeColor string, themeMode string, sidebarBg string, operator string) error {
	pref, err := GetUserPreferenceByUserID(userID)
	if err == nil && pref != nil {
		return db.Model(&UserPreference{}).Where("user_id = ?", userID).Updates(map[string]interface{}{
			"theme_color": themeColor,
			"theme_mode":  themeMode,
			"sidebar_bg":  sidebarBg,
			"modified_by": operator,
		}).Error
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return db.Create(&UserPreference{
		IDModel: IDModel{
			CreatedBy:  operator,
			ModifiedBy: operator,
		},
		UserID:     userID,
		ThemeColor: themeColor,
		ThemeMode:  themeMode,
		SidebarBg:  sidebarBg,
	}).Error
}
