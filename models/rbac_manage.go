package models

import (
	"errors"
	"strings"

	"gorm.io/gorm"
)

func GetRoleByID(id uint) (*RbacRole, error) {
	var role RbacRole
	if err := db.First(&role, id).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func GetRoleByCodeExcludeID(code string, excludeID uint) (*RbacRole, error) {
	var role RbacRole
	query := db.Where("code = ?", code)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}
	if err := query.First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func GetRoles(pageNum int, pageSize int, text string) ([]RbacRole, error) {
	var roles []RbacRole
	query := db.Model(&RbacRole{}).Order("id DESC")
	if strings.TrimSpace(text) != "" {
		like := "%" + strings.TrimSpace(text) + "%"
		query = query.Where("code LIKE ? OR name LIKE ? OR description LIKE ?", like, like, like)
	}
	if pageSize > 0 {
		query = query.Offset(pageNum).Limit(pageSize)
	}
	if err := query.Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

func GetRoleTotal(text string) (int64, error) {
	var total int64
	query := db.Model(&RbacRole{})
	if strings.TrimSpace(text) != "" {
		like := "%" + strings.TrimSpace(text) + "%"
		query = query.Where("code LIKE ? OR name LIKE ? OR description LIKE ?", like, like, like)
	}
	if err := query.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func EditRole(id uint, data map[string]interface{}) error {
	return db.Model(&RbacRole{}).Where("id = ?", id).Updates(data).Error
}

func DeleteRoleByID(id uint) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", id).Delete(&RbacRolePermission{}).Error; err != nil {
			return err
		}
		if err := tx.Where("role_id = ?", id).Delete(&RbacUserRole{}).Error; err != nil {
			return err
		}
		if err := tx.Where("role_id = ?", id).Delete(&RbacGroupRole{}).Error; err != nil {
			return err
		}
		return tx.Delete(&RbacRole{}, id).Error
	})
}

func GetPermissionByID(id uint) (*RbacPermission, error) {
	var permission RbacPermission
	if err := db.First(&permission, id).Error; err != nil {
		return nil, err
	}
	return &permission, nil
}

func GetPermissionByCodeExcludeID(code string, excludeID uint) (*RbacPermission, error) {
	var permission RbacPermission
	query := db.Where("code = ?", code)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}
	if err := query.First(&permission).Error; err != nil {
		return nil, err
	}
	return &permission, nil
}

func GetPermissions(pageNum int, pageSize int, text string, permissionType string) ([]RbacPermission, error) {
	var permissions []RbacPermission
	query := db.Model(&RbacPermission{}).Order("sort ASC, id DESC")
	if strings.TrimSpace(text) != "" {
		like := "%" + strings.TrimSpace(text) + "%"
		query = query.Where("code LIKE ? OR name LIKE ? OR path LIKE ?", like, like, like)
	}
	if strings.TrimSpace(permissionType) != "" {
		query = query.Where("type = ?", strings.TrimSpace(permissionType))
	}
	if pageSize > 0 {
		query = query.Offset(pageNum).Limit(pageSize)
	}
	if err := query.Find(&permissions).Error; err != nil {
		return nil, err
	}
	return permissions, nil
}

func GetPermissionTotal(text string, permissionType string) (int64, error) {
	var total int64
	query := db.Model(&RbacPermission{})
	if strings.TrimSpace(text) != "" {
		like := "%" + strings.TrimSpace(text) + "%"
		query = query.Where("code LIKE ? OR name LIKE ? OR path LIKE ?", like, like, like)
	}
	if strings.TrimSpace(permissionType) != "" {
		query = query.Where("type = ?", strings.TrimSpace(permissionType))
	}
	if err := query.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func EditPermission(id uint, data map[string]interface{}) error {
	return db.Model(&RbacPermission{}).Where("id = ?", id).Updates(data).Error
}

func DeletePermissionByID(id uint) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("permission_id = ?", id).Delete(&RbacRolePermission{}).Error; err != nil {
			return err
		}
		return tx.Delete(&RbacPermission{}, id).Error
	})
}

func GetUserGroupByID(id uint) (*RbacUserGroup, error) {
	var group RbacUserGroup
	if err := db.First(&group, id).Error; err != nil {
		return nil, err
	}
	return &group, nil
}

func GetUserGroupByCodeExcludeID(code string, excludeID uint) (*RbacUserGroup, error) {
	var group RbacUserGroup
	query := db.Where("code = ?", code)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}
	if err := query.First(&group).Error; err != nil {
		return nil, err
	}
	return &group, nil
}

func GetUserGroups(pageNum int, pageSize int, text string) ([]RbacUserGroup, error) {
	var groups []RbacUserGroup
	query := db.Model(&RbacUserGroup{}).Order("id DESC")
	if strings.TrimSpace(text) != "" {
		like := "%" + strings.TrimSpace(text) + "%"
		query = query.Where("code LIKE ? OR name LIKE ? OR description LIKE ?", like, like, like)
	}
	if pageSize > 0 {
		query = query.Offset(pageNum).Limit(pageSize)
	}
	if err := query.Find(&groups).Error; err != nil {
		return nil, err
	}
	return groups, nil
}

func GetUserGroupTotal(text string) (int64, error) {
	var total int64
	query := db.Model(&RbacUserGroup{})
	if strings.TrimSpace(text) != "" {
		like := "%" + strings.TrimSpace(text) + "%"
		query = query.Where("code LIKE ? OR name LIKE ? OR description LIKE ?", like, like, like)
	}
	if err := query.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func EditUserGroup(id uint, data map[string]interface{}) error {
	return db.Model(&RbacUserGroup{}).Where("id = ?", id).Updates(data).Error
}

func DeleteUserGroupByID(id uint) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("group_id = ?", id).Delete(&RbacGroupRole{}).Error; err != nil {
			return err
		}
		if err := tx.Where("group_id = ?", id).Delete(&RbacUserGroupMember{}).Error; err != nil {
			return err
		}
		return tx.Delete(&RbacUserGroup{}, id).Error
	})
}

func IsRoleCodeExists(code string, excludeID uint) (bool, error) {
	_, err := GetRoleByCodeExcludeID(code, excludeID)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return false, err
}

func IsPermissionCodeExists(code string, excludeID uint) (bool, error) {
	_, err := GetPermissionByCodeExcludeID(code, excludeID)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return false, err
}

func IsUserGroupCodeExists(code string, excludeID uint) (bool, error) {
	_, err := GetUserGroupByCodeExcludeID(code, excludeID)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return false, err
}
