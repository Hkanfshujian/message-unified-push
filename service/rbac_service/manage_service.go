package rbac_service

import (
	"errors"
	"strings"

	"ops-message-unified-push/models"
	"ops-message-unified-push/pkg/util"

	"gorm.io/gorm"
)

type RoleManageService struct {
	ID          uint
	Code        string
	Name        string
	Description string
	Status      int
	PageNum     int
	PageSize    int
	Text        string
	Operator    string
}

func (s *RoleManageService) GetList() ([]models.RbacRole, error) {
	return models.GetRoles(s.PageNum, s.PageSize, s.Text)
}

func (s *RoleManageService) Count() (int64, error) {
	return models.GetRoleTotal(s.Text)
}

func (s *RoleManageService) Add() error {
	exist, err := models.IsRoleCodeExists(s.Code, 0)
	if err != nil {
		return err
	}
	if exist {
		return errors.New("角色编码已存在")
	}
	return models.AddRole(&models.RbacRole{
		IDModel: models.IDModel{
			CreatedBy:  s.Operator,
			ModifiedBy: s.Operator,
		},
		Code:        strings.TrimSpace(s.Code),
		Name:        strings.TrimSpace(s.Name),
		Description: strings.TrimSpace(s.Description),
		Status:      s.Status,
	})
}

func (s *RoleManageService) Edit() error {
	exist, err := models.IsRoleCodeExists(s.Code, s.ID)
	if err != nil {
		return err
	}
	if exist {
		return errors.New("角色编码已存在")
	}
	return models.EditRole(s.ID, map[string]interface{}{
		"code":        strings.TrimSpace(s.Code),
		"name":        strings.TrimSpace(s.Name),
		"description": strings.TrimSpace(s.Description),
		"status":      s.Status,
		"modified_by": s.Operator,
	})
}

func (s *RoleManageService) Delete() error {
	if _, err := models.GetRoleByID(s.ID); err != nil {
		return err
	}
	return models.DeleteRoleByID(s.ID)
}

type GroupManageService struct {
	ID          uint
	Code        string
	Name        string
	Description string
	Status      int
	PageNum     int
	PageSize    int
	Text        string
	Operator    string
}

func (s *GroupManageService) GetList() ([]models.RbacUserGroup, error) {
	return models.GetUserGroups(s.PageNum, s.PageSize, s.Text)
}

func (s *GroupManageService) Count() (int64, error) {
	return models.GetUserGroupTotal(s.Text)
}

func (s *GroupManageService) Add() error {
	exist, err := models.IsUserGroupCodeExists(s.Code, 0)
	if err != nil {
		return err
	}
	if exist {
		return errors.New("用户组编码已存在")
	}
	return models.AddUserGroup(&models.RbacUserGroup{
		IDModel: models.IDModel{
			CreatedBy:  s.Operator,
			ModifiedBy: s.Operator,
		},
		Code:        strings.TrimSpace(s.Code),
		Name:        strings.TrimSpace(s.Name),
		Description: strings.TrimSpace(s.Description),
		Status:      s.Status,
	})
}

func (s *GroupManageService) Edit() error {
	exist, err := models.IsUserGroupCodeExists(s.Code, s.ID)
	if err != nil {
		return err
	}
	if exist {
		return errors.New("用户组编码已存在")
	}
	return models.EditUserGroup(s.ID, map[string]interface{}{
		"code":        strings.TrimSpace(s.Code),
		"name":        strings.TrimSpace(s.Name),
		"description": strings.TrimSpace(s.Description),
		"status":      s.Status,
		"modified_by": s.Operator,
	})
}

func (s *GroupManageService) Delete() error {
	if _, err := models.GetUserGroupByID(s.ID); err != nil {
		return err
	}
	return models.DeleteUserGroupByID(s.ID)
}

type PermissionManageService struct {
	ID          uint
	Code        string
	Name        string
	Type        string
	Method      string
	Path        string
	ParentID    uint
	Sort        int
	Status      int
	PageNum     int
	PageSize    int
	Text        string
	FilterType  string
	Description string
	Operator    string
}

func (s *PermissionManageService) GetList() ([]models.RbacPermission, error) {
	return models.GetPermissions(s.PageNum, s.PageSize, s.Text, s.FilterType)
}

func (s *PermissionManageService) Count() (int64, error) {
	return models.GetPermissionTotal(s.Text, s.FilterType)
}

func (s *PermissionManageService) Add() error {
	exist, err := models.IsPermissionCodeExists(s.Code, 0)
	if err != nil {
		return err
	}
	if exist {
		return errors.New("权限编码已存在")
	}
	return models.AddPermission(&models.RbacPermission{
		IDModel: models.IDModel{
			CreatedBy:  s.Operator,
			ModifiedBy: s.Operator,
		},
		Code:     strings.TrimSpace(s.Code),
		Name:     strings.TrimSpace(s.Name),
		Type:     strings.TrimSpace(s.Type),
		Method:   strings.TrimSpace(s.Method),
		Path:     strings.TrimSpace(s.Path),
		ParentID: s.ParentID,
		Sort:     s.Sort,
		Status:   s.Status,
	})
}

func (s *PermissionManageService) Edit() error {
	exist, err := models.IsPermissionCodeExists(s.Code, s.ID)
	if err != nil {
		return err
	}
	if exist {
		return errors.New("权限编码已存在")
	}
	return models.EditPermission(s.ID, map[string]interface{}{
		"code":        strings.TrimSpace(s.Code),
		"name":        strings.TrimSpace(s.Name),
		"type":        strings.TrimSpace(s.Type),
		"method":      strings.TrimSpace(s.Method),
		"path":        strings.TrimSpace(s.Path),
		"parent_id":   s.ParentID,
		"sort":        s.Sort,
		"status":      s.Status,
		"modified_by": s.Operator,
	})
}

func (s *PermissionManageService) Delete() error {
	if _, err := models.GetPermissionByID(s.ID); err != nil {
		return err
	}
	return models.DeletePermissionByID(s.ID)
}

type RelationManageService struct {
	Operator string
}

func (s *RelationManageService) AssignPermissionsToRole(roleID uint, permissionIDs []uint) error {
	if _, err := models.GetRoleByID(roleID); err != nil {
		return err
	}
	return models.SetRolePermissions(roleID, permissionIDs, s.Operator)
}

func (s *RelationManageService) AssignRolesToGroup(groupID uint, roleIDs []uint) error {
	if _, err := models.GetUserGroupByID(groupID); err != nil {
		return err
	}
	return models.SetGroupRoles(groupID, roleIDs, s.Operator)
}

func (s *RelationManageService) AssignMembersToGroup(groupID uint, userIDs []int) error {
	if _, err := models.GetUserGroupByID(groupID); err != nil {
		return err
	}
	return models.SetGroupMembers(groupID, userIDs, s.Operator)
}

func (s *RelationManageService) AssignRolesToUser(userID int, roleIDs []uint) error {
	if _, err := models.GetUserByID(userID); err != nil {
		return err
	}
	return models.SetUserRoles(userID, roleIDs, s.Operator)
}

func (s *RelationManageService) AssignGroupsToUser(userID int, groupIDs []uint) error {
	if _, err := models.GetUserByID(userID); err != nil {
		return err
	}
	return models.SetUserGroups(userID, groupIDs, s.Operator)
}

func (s *RelationManageService) GetPermissionIDsByRole(roleID uint) ([]uint, error) {
	return models.GetPermissionIDsByRoleID(roleID)
}

func (s *RelationManageService) GetRoleIDsByGroup(groupID uint) ([]uint, error) {
	return models.GetRoleIDsByGroupID(groupID)
}

func (s *RelationManageService) GetMemberIDsByGroup(groupID uint) ([]int, error) {
	return models.GetMemberUserIDsByGroupID(groupID)
}

func (s *RelationManageService) GetRoleIDsByUser(userID int) ([]uint, error) {
	return models.GetRoleIDsByUserID(userID)
}

func (s *RelationManageService) GetGroupIDsByUser(userID int) ([]uint, error) {
	return models.GetGroupIDsByUserID(userID)
}

type UserManageService struct {
	ID       int
	Username string
	Password string
	PageNum  int
	PageSize int
	Text     string
	Operator string
}

func (s *UserManageService) GetList() ([]models.Auth, error) {
	return models.GetUsers(s.PageNum, s.PageSize, s.Text)
}

func (s *UserManageService) Count() (int64, error) {
	return models.GetUserTotal(s.Text)
}

func (s *UserManageService) Add() error {
	exist, err := models.IsUsernameExists(strings.TrimSpace(s.Username), 0)
	if err != nil {
		return err
	}
	if exist {
		return errors.New("用户名已存在")
	}
	// 密码使用 MD5 加密存储
	return models.AddUser(strings.TrimSpace(s.Username), util.EncodeMD5(s.Password))
}

func (s *UserManageService) Edit() error {
	if _, err := models.GetUserByID(s.ID); err != nil {
		return err
	}
	exist, err := models.IsUsernameExists(strings.TrimSpace(s.Username), s.ID)
	if err != nil {
		return err
	}
	if exist {
		return errors.New("用户名已存在")
	}
	data := map[string]interface{}{
		"username": strings.TrimSpace(s.Username),
	}
	if strings.TrimSpace(s.Password) != "" {
		data["password"] = util.EncodeMD5(s.Password)
	}
	return models.EditUserByID(s.ID, data)
}

func (s *UserManageService) Delete() error {
	user, err := models.GetUserByID(s.ID)
	if err != nil {
		return err
	}
	if strings.EqualFold(strings.TrimSpace(user.Username), "admin") {
		return errors.New("admin 用户不可删除")
	}
	return models.DeleteUserByID(s.ID)
}

func IsNotFoundErr(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
