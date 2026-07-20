package models

import (
	"errors"

	"gorm.io/gorm"
)

type RbacPermission struct {
	IDModel
	SoftDeleteModel
	Code     string `json:"code" gorm:"type:varchar(100);uniqueIndex;not null"`
	Name     string `json:"name" gorm:"type:varchar(100);not null"`
	Type     string `json:"type" gorm:"type:varchar(20);default:'api'"`
	Method   string `json:"method" gorm:"type:varchar(10);default:''"`
	Path     string `json:"path" gorm:"type:varchar(255);default:''"`
	ParentID uint   `json:"parent_id" gorm:"type:integer;default:0"`
	Sort     int    `json:"sort" gorm:"type:int;default:0"`
	Status   int    `json:"status" gorm:"type:int;default:1"`
}

func archivePermission(tx *gorm.DB, id uint) error {
	if err := archivePermissionRoles(tx, id); err != nil {
		return err
	}
	return tx.Unscoped().Delete(&RbacPermission{}, id).Error
}

func GetPermissionByCode(code string) (*RbacPermission, error) {
	var permission RbacPermission
	if err := db.Where("code = ?", code).First(&permission).Error; err != nil {
		return nil, err
	}
	return &permission, nil
}

func AddPermission(permission *RbacPermission) error {
	var archived RbacPermission
	err := db.Unscoped().Where("code = ?", permission.Code).First(&archived).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return db.Create(permission).Error
	}
	if err != nil {
		return err
	}
	if !archived.DeletedAt.Valid {
		return gorm.ErrDuplicatedKey
	}
	permission.ID = archived.ID
	return db.Unscoped().Model(&archived).Updates(map[string]interface{}{
		"deleted_at": nil, "name": permission.Name, "type": permission.Type,
		"method": permission.Method, "path": permission.Path, "parent_id": permission.ParentID,
		"sort": permission.Sort, "status": permission.Status, "modified_by": permission.ModifiedBy,
	}).Error
}

func AddPermissionIfNotExists(permission *RbacPermission) (*RbacPermission, error) {
	exist, err := GetPermissionByCode(permission.Code)
	if err == nil {
		updates := map[string]interface{}{}
		if exist.Name != permission.Name {
			updates["name"] = permission.Name
		}
		if exist.Type != permission.Type {
			updates["type"] = permission.Type
		}
		if exist.Method != permission.Method {
			updates["method"] = permission.Method
		}
		if exist.Path != permission.Path {
			updates["path"] = permission.Path
		}
		if exist.Sort != permission.Sort {
			updates["sort"] = permission.Sort
		}
		if exist.Status != permission.Status {
			updates["status"] = permission.Status
		}
		if len(updates) > 0 {
			updates["modified_by"] = permission.ModifiedBy
			if err = db.Model(exist).Updates(updates).Error; err != nil {
				return nil, err
			}
			return GetPermissionByCode(permission.Code)
		}
		return exist, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if err = AddPermission(permission); err != nil {
		return nil, err
	}
	return permission, nil
}
