package models

import "gorm.io/gorm"

// RBAC join rows are replace-set associations, so each set is rebuilt atomically.

func SetRolePermissions(roleID uint, permissionIDs []uint, operator string) error {
	return WithTransaction(func(tx *gorm.DB) error {
		return replaceRolePermissions(tx, roleID, permissionIDs, operator)
	})
}

func SetGroupRoles(groupID uint, roleIDs []uint, operator string) error {
	return WithTransaction(func(tx *gorm.DB) error {
		return replaceGroupRoles(tx, groupID, roleIDs, operator)
	})
}

func SetGroupMembers(groupID uint, userIDs []int, operator string) error {
	return WithTransaction(func(tx *gorm.DB) error {
		return replaceGroupMembers(tx, groupID, userIDs, operator)
	})
}

func SetUserRoles(userID int, roleIDs []uint, operator string) error {
	return WithTransaction(func(tx *gorm.DB) error {
		return replaceUserRoles(tx, userID, roleIDs, operator)
	})
}

func SetUserGroups(userID int, groupIDs []uint, operator string) error {
	return WithTransaction(func(tx *gorm.DB) error {
		return replaceUserGroups(tx, userID, groupIDs, operator)
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
