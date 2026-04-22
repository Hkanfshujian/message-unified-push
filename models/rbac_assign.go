package models

import "gorm.io/gorm"

func SetRolePermissions(roleID uint, permissionIDs []uint, operator string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", roleID).Delete(&RbacRolePermission{}).Error; err != nil {
			return err
		}
		for _, permissionID := range permissionIDs {
			if err := tx.Create(&RbacRolePermission{
				IDModel: IDModel{
					CreatedBy:  operator,
					ModifiedBy: operator,
				},
				RoleID:       roleID,
				PermissionID: permissionID,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func SetGroupRoles(groupID uint, roleIDs []uint, operator string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("group_id = ?", groupID).Delete(&RbacGroupRole{}).Error; err != nil {
			return err
		}
		for _, roleID := range roleIDs {
			if err := tx.Create(&RbacGroupRole{
				IDModel: IDModel{
					CreatedBy:  operator,
					ModifiedBy: operator,
				},
				GroupID: groupID,
				RoleID:  roleID,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func SetGroupMembers(groupID uint, userIDs []int, operator string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("group_id = ?", groupID).Delete(&RbacUserGroupMember{}).Error; err != nil {
			return err
		}
		for _, userID := range userIDs {
			if err := tx.Create(&RbacUserGroupMember{
				IDModel: IDModel{
					CreatedBy:  operator,
					ModifiedBy: operator,
				},
				UserID:  userID,
				GroupID: groupID,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func SetUserRoles(userID int, roleIDs []uint, operator string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&RbacUserRole{}).Error; err != nil {
			return err
		}
		for _, roleID := range roleIDs {
			if err := tx.Create(&RbacUserRole{
				IDModel: IDModel{
					CreatedBy:  operator,
					ModifiedBy: operator,
				},
				UserID: userID,
				RoleID: roleID,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func SetUserGroups(userID int, groupIDs []uint, operator string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&RbacUserGroupMember{}).Error; err != nil {
			return err
		}
		for _, groupID := range groupIDs {
			if err := tx.Create(&RbacUserGroupMember{
				IDModel: IDModel{
					CreatedBy:  operator,
					ModifiedBy: operator,
				},
				UserID:  userID,
				GroupID: groupID,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func GetPermissionIDsByRoleID(roleID uint) ([]uint, error) {
	var list []RbacRolePermission
	if err := db.Where("role_id = ?", roleID).Find(&list).Error; err != nil {
		return nil, err
	}
	ids := make([]uint, 0, len(list))
	for _, item := range list {
		ids = append(ids, item.PermissionID)
	}
	return ids, nil
}

func GetRoleIDsByGroupID(groupID uint) ([]uint, error) {
	var list []RbacGroupRole
	if err := db.Where("group_id = ?", groupID).Find(&list).Error; err != nil {
		return nil, err
	}
	ids := make([]uint, 0, len(list))
	for _, item := range list {
		ids = append(ids, item.RoleID)
	}
	return ids, nil
}

func GetMemberUserIDsByGroupID(groupID uint) ([]int, error) {
	var list []RbacUserGroupMember
	if err := db.Where("group_id = ?", groupID).Find(&list).Error; err != nil {
		return nil, err
	}
	ids := make([]int, 0, len(list))
	for _, item := range list {
		ids = append(ids, item.UserID)
	}
	return ids, nil
}

func GetRoleIDsByUserID(userID int) ([]uint, error) {
	var list []RbacUserRole
	if err := db.Where("user_id = ?", userID).Find(&list).Error; err != nil {
		return nil, err
	}
	ids := make([]uint, 0, len(list))
	for _, item := range list {
		ids = append(ids, item.RoleID)
	}
	return ids, nil
}

func GetGroupIDsByUserID(userID int) ([]uint, error) {
	var list []RbacUserGroupMember
	if err := db.Where("user_id = ?", userID).Find(&list).Error; err != nil {
		return nil, err
	}
	ids := make([]uint, 0, len(list))
	for _, item := range list {
		ids = append(ids, item.GroupID)
	}
	return ids, nil
}
