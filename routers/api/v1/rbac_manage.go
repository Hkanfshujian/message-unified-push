package v1

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"ops-message-unified-push/pkg/app"
	"ops-message-unified-push/pkg/e"
	"ops-message-unified-push/pkg/util"
	"ops-message-unified-push/service/rbac_service"

	"github.com/gin-gonic/gin"
)

type AddRoleReq struct {
	Code        string `json:"code" validate:"required,max=100,min=2" label:"角色编码"`
	Name        string `json:"name" validate:"required,max=100,min=1" label:"角色名称"`
	Description string `json:"description" validate:"max=255" label:"角色描述"`
	Status      int    `json:"status" validate:"omitempty,oneof=0 1" label:"角色状态"`
}

type EditRoleReq struct {
	ID          uint   `json:"id" validate:"required" label:"角色ID"`
	Code        string `json:"code" validate:"required,max=100,min=2" label:"角色编码"`
	Name        string `json:"name" validate:"required,max=100,min=1" label:"角色名称"`
	Description string `json:"description" validate:"max=255" label:"角色描述"`
	Status      int    `json:"status" validate:"omitempty,oneof=0 1" label:"角色状态"`
}

type DeleteRoleReq struct {
	ID uint `json:"id" validate:"required" label:"角色ID"`
}

func GetRbacRoles(c *gin.Context) {
	appG := app.Gin{C: c}
	text := c.Query("text")
	offset, limit := util.GetPageSize(c)
	service := rbac_service.RoleManageService{
		PageNum:  offset,
		PageSize: limit,
		Text:     text,
	}
	list, err := service.GetList()
	if err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("获取角色列表失败：%s", err.Error()), nil)
		return
	}
	total, err := service.Count()
	if err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("获取角色总数失败：%s", err.Error()), nil)
		return
	}
	appG.CResponse(http.StatusOK, "获取角色列表成功", map[string]interface{}{
		"lists": list,
		"total": total,
	})
}

func AddRbacRole(c *gin.Context) {
	var req AddRoleReq
	appG := app.Gin{C: c}
	errCode, errMsg := app.BindJsonAndPlayValid(c, &req)
	if errCode != e.SUCCESS {
		appG.CResponse(errCode, errMsg, nil)
		return
	}
	if req.Status != 0 {
		req.Status = 1
	}
	service := rbac_service.RoleManageService{
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
		Status:      req.Status,
		Operator:    app.GetCurrentUserName(c),
	}
	if err := service.Add(); err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("添加角色失败：%s", err.Error()), nil)
		return
	}
	appG.CResponse(http.StatusOK, "添加角色成功", nil)
}

func EditRbacRole(c *gin.Context) {
	var req EditRoleReq
	appG := app.Gin{C: c}
	errCode, errMsg := app.BindJsonAndPlayValid(c, &req)
	if errCode != e.SUCCESS {
		appG.CResponse(errCode, errMsg, nil)
		return
	}
	service := rbac_service.RoleManageService{
		ID:          req.ID,
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
		Status:      req.Status,
		Operator:    app.GetCurrentUserName(c),
	}
	if err := service.Edit(); err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("编辑角色失败：%s", err.Error()), nil)
		return
	}
	appG.CResponse(http.StatusOK, "编辑角色成功", nil)
}

func DeleteRbacRole(c *gin.Context) {
	var req DeleteRoleReq
	appG := app.Gin{C: c}
	errCode, errMsg := app.BindJsonAndPlayValid(c, &req)
	if errCode != e.SUCCESS {
		appG.CResponse(errCode, errMsg, nil)
		return
	}
	service := rbac_service.RoleManageService{ID: req.ID}
	if err := service.Delete(); err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("删除角色失败：%s", err.Error()), nil)
		return
	}
	appG.CResponse(http.StatusOK, "删除角色成功", nil)
}

type AddGroupReq struct {
	Code        string `json:"code" validate:"required,max=100,min=2" label:"用户组编码"`
	Name        string `json:"name" validate:"required,max=100,min=1" label:"用户组名称"`
	Description string `json:"description" validate:"max=255" label:"用户组描述"`
	Status      int    `json:"status" validate:"omitempty,oneof=0 1" label:"用户组状态"`
}

type EditGroupReq struct {
	ID          uint   `json:"id" validate:"required" label:"用户组ID"`
	Code        string `json:"code" validate:"required,max=100,min=2" label:"用户组编码"`
	Name        string `json:"name" validate:"required,max=100,min=1" label:"用户组名称"`
	Description string `json:"description" validate:"max=255" label:"用户组描述"`
	Status      int    `json:"status" validate:"omitempty,oneof=0 1" label:"用户组状态"`
}

type DeleteGroupReq struct {
	ID uint `json:"id" validate:"required" label:"用户组ID"`
}

func GetRbacGroups(c *gin.Context) {
	appG := app.Gin{C: c}
	text := c.Query("text")
	offset, limit := util.GetPageSize(c)
	service := rbac_service.GroupManageService{
		PageNum:  offset,
		PageSize: limit,
		Text:     text,
	}
	list, err := service.GetList()
	if err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("获取用户组列表失败：%s", err.Error()), nil)
		return
	}
	total, err := service.Count()
	if err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("获取用户组总数失败：%s", err.Error()), nil)
		return
	}
	appG.CResponse(http.StatusOK, "获取用户组列表成功", map[string]interface{}{
		"lists": list,
		"total": total,
	})
}

func AddRbacGroup(c *gin.Context) {
	var req AddGroupReq
	appG := app.Gin{C: c}
	errCode, errMsg := app.BindJsonAndPlayValid(c, &req)
	if errCode != e.SUCCESS {
		appG.CResponse(errCode, errMsg, nil)
		return
	}
	if req.Status != 0 {
		req.Status = 1
	}
	service := rbac_service.GroupManageService{
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
		Status:      req.Status,
		Operator:    app.GetCurrentUserName(c),
	}
	if err := service.Add(); err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("添加用户组失败：%s", err.Error()), nil)
		return
	}
	appG.CResponse(http.StatusOK, "添加用户组成功", nil)
}

func EditRbacGroup(c *gin.Context) {
	var req EditGroupReq
	appG := app.Gin{C: c}
	errCode, errMsg := app.BindJsonAndPlayValid(c, &req)
	if errCode != e.SUCCESS {
		appG.CResponse(errCode, errMsg, nil)
		return
	}
	service := rbac_service.GroupManageService{
		ID:          req.ID,
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
		Status:      req.Status,
		Operator:    app.GetCurrentUserName(c),
	}
	if err := service.Edit(); err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("编辑用户组失败：%s", err.Error()), nil)
		return
	}
	appG.CResponse(http.StatusOK, "编辑用户组成功", nil)
}

func DeleteRbacGroup(c *gin.Context) {
	var req DeleteGroupReq
	appG := app.Gin{C: c}
	errCode, errMsg := app.BindJsonAndPlayValid(c, &req)
	if errCode != e.SUCCESS {
		appG.CResponse(errCode, errMsg, nil)
		return
	}
	service := rbac_service.GroupManageService{ID: req.ID}
	if err := service.Delete(); err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("删除用户组失败：%s", err.Error()), nil)
		return
	}
	appG.CResponse(http.StatusOK, "删除用户组成功", nil)
}

type AddPermissionReq struct {
	Code     string `json:"code" validate:"required,max=100,min=2" label:"权限编码"`
	Name     string `json:"name" validate:"required,max=100,min=1" label:"权限名称"`
	Type     string `json:"type" validate:"required,oneof=menu action api" label:"权限类型"`
	Method   string `json:"method" validate:"max=10" label:"请求方法"`
	Path     string `json:"path" validate:"max=255" label:"请求路径"`
	ParentID uint   `json:"parent_id" label:"父级ID"`
	Sort     int    `json:"sort" label:"排序"`
	Status   int    `json:"status" validate:"omitempty,oneof=0 1" label:"权限状态"`
}

type EditPermissionReq struct {
	ID       uint   `json:"id" validate:"required" label:"权限ID"`
	Code     string `json:"code" validate:"required,max=100,min=2" label:"权限编码"`
	Name     string `json:"name" validate:"required,max=100,min=1" label:"权限名称"`
	Type     string `json:"type" validate:"required,oneof=menu action api" label:"权限类型"`
	Method   string `json:"method" validate:"max=10" label:"请求方法"`
	Path     string `json:"path" validate:"max=255" label:"请求路径"`
	ParentID uint   `json:"parent_id" label:"父级ID"`
	Sort     int    `json:"sort" label:"排序"`
	Status   int    `json:"status" validate:"omitempty,oneof=0 1" label:"权限状态"`
}

func GetRbacPermissions(c *gin.Context) {
	appG := app.Gin{C: c}
	text := c.Query("text")
	typeFilter := c.Query("type")
	offset, limit := util.GetPageSize(c)
	service := rbac_service.PermissionManageService{
		PageNum:    offset,
		PageSize:   limit,
		Text:       text,
		FilterType: typeFilter,
	}
	list, err := service.GetList()
	if err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("获取权限列表失败：%s", err.Error()), nil)
		return
	}
	total, err := service.Count()
	if err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("获取权限总数失败：%s", err.Error()), nil)
		return
	}
	appG.CResponse(http.StatusOK, "获取权限列表成功", map[string]interface{}{
		"lists": list,
		"total": total,
	})
}

func AddRbacPermission(c *gin.Context) {
	var req AddPermissionReq
	appG := app.Gin{C: c}
	errCode, errMsg := app.BindJsonAndPlayValid(c, &req)
	if errCode != e.SUCCESS {
		appG.CResponse(errCode, errMsg, nil)
		return
	}
	if req.Status != 0 {
		req.Status = 1
	}
	service := rbac_service.PermissionManageService{
		Code:     req.Code,
		Name:     req.Name,
		Type:     req.Type,
		Method:   strings.ToUpper(req.Method),
		Path:     req.Path,
		ParentID: req.ParentID,
		Sort:     req.Sort,
		Status:   req.Status,
		Operator: app.GetCurrentUserName(c),
	}
	if err := service.Add(); err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("添加权限失败：%s", err.Error()), nil)
		return
	}
	appG.CResponse(http.StatusOK, "添加权限成功", nil)
}

func EditRbacPermission(c *gin.Context) {
	var req EditPermissionReq
	appG := app.Gin{C: c}
	errCode, errMsg := app.BindJsonAndPlayValid(c, &req)
	if errCode != e.SUCCESS {
		appG.CResponse(errCode, errMsg, nil)
		return
	}
	service := rbac_service.PermissionManageService{
		ID:       req.ID,
		Code:     req.Code,
		Name:     req.Name,
		Type:     req.Type,
		Method:   strings.ToUpper(req.Method),
		Path:     req.Path,
		ParentID: req.ParentID,
		Sort:     req.Sort,
		Status:   req.Status,
		Operator: app.GetCurrentUserName(c),
	}
	if err := service.Edit(); err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("编辑权限失败：%s", err.Error()), nil)
		return
	}
	appG.CResponse(http.StatusOK, "编辑权限成功", nil)
}

type AssignRolePermissionsReq struct {
	RoleID        uint   `json:"role_id" validate:"required" label:"角色ID"`
	PermissionIDs []uint `json:"permission_ids" label:"权限ID列表"`
}

type AssignGroupRolesReq struct {
	GroupID uint   `json:"group_id" validate:"required" label:"用户组ID"`
	RoleIDs []uint `json:"role_ids" label:"角色ID列表"`
}

type AssignGroupMembersReq struct {
	GroupID uint  `json:"group_id" validate:"required" label:"用户组ID"`
	UserIDs []int `json:"user_ids" label:"用户ID列表"`
}

type AssignUserRolesReq struct {
	UserID  int    `json:"user_id" validate:"required" label:"用户ID"`
	RoleIDs []uint `json:"role_ids" label:"角色ID列表"`
}

type AssignUserGroupsReq struct {
	UserID   int    `json:"user_id" validate:"required" label:"用户ID"`
	GroupIDs []uint `json:"group_ids" label:"用户组ID列表"`
}

type AddManageUserReq struct {
	Username string `json:"username" validate:"required,min=3,max=50" label:"用户名"`
	Password string `json:"passwd" validate:"required,min=6,max=50" label:"密码"`
}

type EditManageUserReq struct {
	ID       int    `json:"id" validate:"required" label:"用户ID"`
	Username string `json:"username" validate:"required,min=3,max=50" label:"用户名"`
	Password string `json:"passwd" validate:"omitempty,min=6,max=50" label:"密码"`
}

type DeleteManageUserReq struct {
	ID int `json:"id" validate:"required" label:"用户ID"`
}

func AssignPermissionsToRole(c *gin.Context) {
	var req AssignRolePermissionsReq
	appG := app.Gin{C: c}
	errCode, errMsg := app.BindJsonAndPlayValid(c, &req)
	if errCode != e.SUCCESS {
		appG.CResponse(errCode, errMsg, nil)
		return
	}
	service := rbac_service.RelationManageService{Operator: app.GetCurrentUserName(c)}
	if err := service.AssignPermissionsToRole(req.RoleID, req.PermissionIDs); err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("角色授权失败：%s", err.Error()), nil)
		return
	}
	appG.CResponse(http.StatusOK, "角色授权成功", nil)
}

func AssignRolesToGroup(c *gin.Context) {
	var req AssignGroupRolesReq
	appG := app.Gin{C: c}
	errCode, errMsg := app.BindJsonAndPlayValid(c, &req)
	if errCode != e.SUCCESS {
		appG.CResponse(errCode, errMsg, nil)
		return
	}
	service := rbac_service.RelationManageService{Operator: app.GetCurrentUserName(c)}
	if err := service.AssignRolesToGroup(req.GroupID, req.RoleIDs); err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("用户组角色授权失败：%s", err.Error()), nil)
		return
	}
	appG.CResponse(http.StatusOK, "用户组角色授权成功", nil)
}

func AssignMembersToGroup(c *gin.Context) {
	var req AssignGroupMembersReq
	appG := app.Gin{C: c}
	errCode, errMsg := app.BindJsonAndPlayValid(c, &req)
	if errCode != e.SUCCESS {
		appG.CResponse(errCode, errMsg, nil)
		return
	}
	service := rbac_service.RelationManageService{Operator: app.GetCurrentUserName(c)}
	if err := service.AssignMembersToGroup(req.GroupID, req.UserIDs); err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("用户组成员授权失败：%s", err.Error()), nil)
		return
	}
	appG.CResponse(http.StatusOK, "用户组成员授权成功", nil)
}

func AssignRolesToUser(c *gin.Context) {
	var req AssignUserRolesReq
	appG := app.Gin{C: c}
	errCode, errMsg := app.BindJsonAndPlayValid(c, &req)
	if errCode != e.SUCCESS {
		appG.CResponse(errCode, errMsg, nil)
		return
	}
	service := rbac_service.RelationManageService{Operator: app.GetCurrentUserName(c)}
	if err := service.AssignRolesToUser(req.UserID, req.RoleIDs); err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("用户角色授权失败：%s", err.Error()), nil)
		return
	}
	appG.CResponse(http.StatusOK, "用户角色授权成功", nil)
}

func AssignGroupsToUser(c *gin.Context) {
	var req AssignUserGroupsReq
	appG := app.Gin{C: c}
	errCode, errMsg := app.BindJsonAndPlayValid(c, &req)
	if errCode != e.SUCCESS {
		appG.CResponse(errCode, errMsg, nil)
		return
	}
	service := rbac_service.RelationManageService{Operator: app.GetCurrentUserName(c)}
	if err := service.AssignGroupsToUser(req.UserID, req.GroupIDs); err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("用户组授权失败：%s", err.Error()), nil)
		return
	}
	appG.CResponse(http.StatusOK, "用户组授权成功", nil)
}

func GetRolePermissionIDs(c *gin.Context) {
	appG := app.Gin{C: c}
	roleID, err := parseUintFromQuery(c, "role_id")
	if err != nil {
		appG.CResponse(http.StatusBadRequest, "role_id 参数错误", nil)
		return
	}
	service := rbac_service.RelationManageService{}
	ids, err := service.GetPermissionIDsByRole(roleID)
	if err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("获取角色权限失败：%s", err.Error()), nil)
		return
	}
	appG.CResponse(http.StatusOK, "获取角色权限成功", map[string]interface{}{"permission_ids": ids})
}

func GetGroupRoleIDs(c *gin.Context) {
	appG := app.Gin{C: c}
	groupID, err := parseUintFromQuery(c, "group_id")
	if err != nil {
		appG.CResponse(http.StatusBadRequest, "group_id 参数错误", nil)
		return
	}
	service := rbac_service.RelationManageService{}
	ids, err := service.GetRoleIDsByGroup(groupID)
	if err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("获取用户组角色失败：%s", err.Error()), nil)
		return
	}
	appG.CResponse(http.StatusOK, "获取用户组角色成功", map[string]interface{}{"role_ids": ids})
}

func GetGroupMemberIDs(c *gin.Context) {
	appG := app.Gin{C: c}
	groupID, err := parseUintFromQuery(c, "group_id")
	if err != nil {
		appG.CResponse(http.StatusBadRequest, "group_id 参数错误", nil)
		return
	}
	service := rbac_service.RelationManageService{}
	ids, err := service.GetMemberIDsByGroup(groupID)
	if err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("获取用户组成员失败：%s", err.Error()), nil)
		return
	}
	appG.CResponse(http.StatusOK, "获取用户组成员成功", map[string]interface{}{"user_ids": ids})
}

func GetRbacUsers(c *gin.Context) {
	appG := app.Gin{C: c}
	text := c.Query("text")
	offset, limit := util.GetPageSize(c)
	service := rbac_service.UserManageService{
		PageNum:  offset,
		PageSize: limit,
		Text:     text,
	}
	list, err := service.GetList()
	if err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("获取用户列表失败：%s", err.Error()), nil)
		return
	}
	total, err := service.Count()
	if err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("获取用户总数失败：%s", err.Error()), nil)
		return
	}
	appG.CResponse(http.StatusOK, "获取用户列表成功", map[string]interface{}{
		"lists": list,
		"total": total,
	})
}

func GetManageUsers(c *gin.Context) {
	appG := app.Gin{C: c}
	text := c.Query("text")
	offset, limit := util.GetPageSize(c)
	service := rbac_service.UserManageService{
		PageNum:  offset,
		PageSize: limit,
		Text:     text,
	}
	list, err := service.GetList()
	if err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("获取用户列表失败：%s", err.Error()), nil)
		return
	}
	total, err := service.Count()
	if err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("获取用户总数失败：%s", err.Error()), nil)
		return
	}
	appG.CResponse(http.StatusOK, "获取用户列表成功", map[string]interface{}{
		"lists": list,
		"total": total,
	})
}

func AddManageUser(c *gin.Context) {
	var req AddManageUserReq
	appG := app.Gin{C: c}
	errCode, errMsg := app.BindJsonAndPlayValid(c, &req)
	if errCode != e.SUCCESS {
		appG.CResponse(errCode, errMsg, nil)
		return
	}
	service := rbac_service.UserManageService{
		Username: req.Username,
		Password: req.Password,
		Operator: app.GetCurrentUserName(c),
	}
	if err := service.Add(); err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("新增用户失败：%s", err.Error()), nil)
		return
	}
	appG.CResponse(http.StatusOK, "新增用户成功", nil)
}

func EditManageUser(c *gin.Context) {
	var req EditManageUserReq
	appG := app.Gin{C: c}
	errCode, errMsg := app.BindJsonAndPlayValid(c, &req)
	if errCode != e.SUCCESS {
		appG.CResponse(errCode, errMsg, nil)
		return
	}
	service := rbac_service.UserManageService{
		ID:       req.ID,
		Username: req.Username,
		Password: req.Password,
		Operator: app.GetCurrentUserName(c),
	}
	if err := service.Edit(); err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("编辑用户失败：%s", err.Error()), nil)
		return
	}
	appG.CResponse(http.StatusOK, "编辑用户成功", nil)
}

func DeleteManageUser(c *gin.Context) {
	var req DeleteManageUserReq
	appG := app.Gin{C: c}
	errCode, errMsg := app.BindJsonAndPlayValid(c, &req)
	if errCode != e.SUCCESS {
		appG.CResponse(errCode, errMsg, nil)
		return
	}
	service := rbac_service.UserManageService{ID: req.ID}
	if err := service.Delete(); err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("删除用户失败：%s", err.Error()), nil)
		return
	}
	appG.CResponse(http.StatusOK, "删除用户成功", nil)
}

func GetUserRoleIDs(c *gin.Context) {
	appG := app.Gin{C: c}
	userID, err := parseIntFromQuery(c, "user_id")
	if err != nil {
		appG.CResponse(http.StatusBadRequest, "user_id 参数错误", nil)
		return
	}
	service := rbac_service.RelationManageService{}
	ids, err := service.GetRoleIDsByUser(userID)
	if err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("获取用户角色失败：%s", err.Error()), nil)
		return
	}
	appG.CResponse(http.StatusOK, "获取用户角色成功", map[string]interface{}{"role_ids": ids})
}

func GetUserGroupIDs(c *gin.Context) {
	appG := app.Gin{C: c}
	userID, err := parseIntFromQuery(c, "user_id")
	if err != nil {
		appG.CResponse(http.StatusBadRequest, "user_id 参数错误", nil)
		return
	}
	service := rbac_service.RelationManageService{}
	ids, err := service.GetGroupIDsByUser(userID)
	if err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("获取用户组失败：%s", err.Error()), nil)
		return
	}
	appG.CResponse(http.StatusOK, "获取用户组成功", map[string]interface{}{"group_ids": ids})
}

func parseUintFromQuery(c *gin.Context, key string) (uint, error) {
	value := c.Query(key)
	id, err := strconv.Atoi(value)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid id")
	}
	return uint(id), nil
}

func parseIntFromQuery(c *gin.Context, key string) (int, error) {
	value := c.Query(key)
	id, err := strconv.Atoi(value)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid id")
	}
	return id, nil
}
