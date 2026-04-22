package rbac_service

import "ops-message-unified-push/models"

type UserAuthzService struct {
	UserName string
}

func (s *UserAuthzService) GetPermissionCodes() ([]string, error) {
	user, err := models.GetUserByUsername(s.UserName)
	if err != nil {
		return nil, err
	}
	permissions, err := models.GetPermissionCodesByUserID(user.ID)
	if err != nil {
		return nil, err
	}
	return appendBuiltinProfilePermissions(permissions), nil
}

func (s *UserAuthzService) HasPermission(code string) (bool, error) {
	permissions, err := s.GetPermissionCodes()
	if err != nil {
		return false, err
	}
	return containsRole(permissions, code), nil
}

func (s *UserAuthzService) HasAnyPermission(codes []string) (bool, error) {
	permissions, err := s.GetPermissionCodes()
	if err != nil {
		return false, err
	}
	permissionMap := make(map[string]struct{}, len(permissions))
	for _, permission := range permissions {
		permissionMap[permission] = struct{}{}
	}
	for _, code := range codes {
		if _, ok := permissionMap[code]; ok {
			return true, nil
		}
	}
	return false, nil
}

func (s *UserAuthzService) GetCurrentUserPermissions() (map[string]interface{}, error) {
	user, err := models.GetUserByUsername(s.UserName)
	if err != nil {
		return nil, err
	}

	permissions, err := models.GetPermissionCodesByUserID(user.ID)
	if err != nil {
		return nil, err
	}
	permissions = appendBuiltinProfilePermissions(permissions)

	roles, err := models.GetRoleCodesByUserID(user.ID)
	if err != nil {
		return nil, err
	}

	groups, err := models.GetGroupCodesByUserID(user.ID)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"user_id":        user.ID,
		"username":       user.Username,
		"roles":          roles,
		"groups":         groups,
		"permissions":    permissions,
		"is_super_admin": containsRole(roles, "super_admin"),
	}, nil
}

func containsRole(roles []string, target string) bool {
	for _, role := range roles {
		if role == target {
			return true
		}
	}
	return false
}

func appendBuiltinProfilePermissions(permissions []string) []string {
	permissionMap := make(map[string]struct{}, len(permissions)+2)
	result := make([]string, 0, len(permissions)+2)
	for _, code := range permissions {
		if _, ok := permissionMap[code]; ok {
			continue
		}
		permissionMap[code] = struct{}{}
		result = append(result, code)
	}
	for _, code := range []string{"profile:settings:view", "profile:settings:edit"} {
		if _, ok := permissionMap[code]; ok {
			continue
		}
		permissionMap[code] = struct{}{}
		result = append(result, code)
	}
	return result
}
