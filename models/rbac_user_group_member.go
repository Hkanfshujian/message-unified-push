package models

import (
	"errors"

	"gorm.io/gorm"
)

type RbacUserGroupMember struct {
	IDModel
	SoftDeleteModel
	UserID  int  `json:"user_id" gorm:"index:idx_user_group_member,unique;not null"`
	GroupID uint `json:"group_id" gorm:"index:idx_user_group_member,unique;not null"`
}

func archiveGroupMembers(tx *gorm.DB, groupID uint) error {
	return tx.Unscoped().Where("group_id = ?", groupID).Delete(&RbacUserGroupMember{}).Error
}

func archiveUserGroups(tx *gorm.DB, userID int) error {
	return tx.Unscoped().Where("user_id = ?", userID).Delete(&RbacUserGroupMember{}).Error
}

func replaceGroupMembers(tx *gorm.DB, groupID uint, userIDs []int, operator string) error {
	if err := archiveGroupMembers(tx, groupID); err != nil {
		return err
	}
	for _, userID := range userIDs {
		if err := restoreGroupMember(tx, userID, groupID, operator); err != nil {
			return err
		}
	}
	return nil
}

func replaceUserGroups(tx *gorm.DB, userID int, groupIDs []uint, operator string) error {
	if err := archiveUserGroups(tx, userID); err != nil {
		return err
	}
	for _, groupID := range groupIDs {
		if err := restoreGroupMember(tx, userID, groupID, operator); err != nil {
			return err
		}
	}
	return nil
}

func restoreGroupMember(tx *gorm.DB, userID int, groupID uint, operator string) error {
	var relation RbacUserGroupMember
	err := tx.Unscoped().Where("user_id = ? AND group_id = ?", userID, groupID).First(&relation).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		relation = RbacUserGroupMember{
			IDModel: IDModel{CreatedBy: operator, ModifiedBy: operator},
			UserID:  userID, GroupID: groupID,
		}
		return tx.Create(&relation).Error
	}
	if err != nil {
		return err
	}
	return tx.Unscoped().Model(&relation).Updates(map[string]interface{}{
		"deleted_at": nil, "modified_by": operator,
	}).Error
}

func AssignUserToGroupIfNotExists(userID int, groupID uint, operator string) error {
	return restoreGroupMember(db, userID, groupID, operator)
}
