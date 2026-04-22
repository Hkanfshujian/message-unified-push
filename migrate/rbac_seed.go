package migrate

import (
	"ops-message-unified-push/models"

	"github.com/sirupsen/logrus"
)

type permissionSeed struct {
	Code   string
	Name   string
	Type   string
	Method string
	Path   string
	Sort   int
}

var defaultPermissionSeeds = []permissionSeed{
	{Code: "dashboard:view", Name: "查看数据统计", Type: "menu", Path: "/", Sort: 1},
	{Code: "message:cron:view", Name: "查看定时消息", Type: "menu", Path: "/cronmessages", Sort: 2},
	{Code: "message:cron:add", Name: "新增定时消息", Type: "action", Method: "POST", Path: "/api/v1/cronmessages/addone", Sort: 3},
	{Code: "message:cron:edit", Name: "编辑定时消息", Type: "action", Method: "POST", Path: "/api/v1/cronmessages/edit", Sort: 4},
	{Code: "message:cron:delete", Name: "删除定时消息", Type: "action", Method: "POST", Path: "/api/v1/cronmessages/delete", Sort: 5},
	{Code: "message:cron:sendnow", Name: "立即发送定时消息", Type: "action", Method: "POST", Path: "/api/v1/cronmessages/sendnow", Sort: 6},
	{Code: "message:template:view", Name: "查看模板管理", Type: "menu", Path: "/templates", Sort: 7},
	{Code: "message:template:add", Name: "新增模板", Type: "action", Method: "POST", Path: "/api/v1/templates/add", Sort: 8},
	{Code: "message:template:edit", Name: "编辑模板", Type: "action", Method: "POST", Path: "/api/v1/templates/edit", Sort: 9},
	{Code: "message:template:delete", Name: "删除模板", Type: "action", Method: "POST", Path: "/api/v1/templates/delete", Sort: 10},
	{Code: "message:template:preview", Name: "预览模板", Type: "action", Method: "POST", Path: "/api/v1/templates/preview", Sort: 11},
	{Code: "message:template:instance", Name: "管理模板实例", Type: "action", Method: "POST", Path: "/api/v1/templates/ins/addone", Sort: 12},
	{Code: "message:sendways:view", Name: "查看渠道管理", Type: "menu", Path: "/sendways", Sort: 13},
	{Code: "message:sendways:add", Name: "新增渠道", Type: "action", Method: "POST", Path: "/api/v1/sendways/add", Sort: 14},
	{Code: "message:sendways:edit", Name: "编辑渠道", Type: "action", Method: "POST", Path: "/api/v1/sendways/edit", Sort: 15},
	{Code: "message:sendways:delete", Name: "删除渠道", Type: "action", Method: "POST", Path: "/api/v1/sendways/delete", Sort: 16},
	{Code: "message:sendways:test", Name: "测试渠道", Type: "action", Method: "POST", Path: "/api/v1/sendways/test", Sort: 17},
	{Code: "message:sendlogs:view", Name: "查看日志管理", Type: "menu", Path: "/sendlogs", Sort: 18},
	{Code: "system:settings:view", Name: "查看系统设置", Type: "menu", Path: "/settings", Sort: 19},
	{Code: "system:settings:edit", Name: "编辑系统设置", Type: "action", Method: "POST", Path: "/api/v1/settings/set", Sort: 20},
	{Code: "profile:settings:view", Name: "查看个人设置", Type: "menu", Path: "/profile/settings", Sort: 21},
	{Code: "profile:settings:edit", Name: "编辑个人设置", Type: "action", Method: "POST", Path: "/api/v1/profile/password", Sort: 22},
	{Code: "system:loginlogs:view", Name: "查看登录日志", Type: "api", Method: "GET", Path: "/api/v1/loginlogs/recent", Sort: 23},
	{Code: "system:rbac:view", Name: "查看系统管理", Type: "menu", Path: "/system", Sort: 22},
	{Code: "system:rbac:role", Name: "角色管理", Type: "menu", Path: "/system/roles", Sort: 23},
	{Code: "system:rbac:group", Name: "用户组管理", Type: "menu", Path: "/system/groups", Sort: 24},
	{Code: "system:rbac:permission", Name: "权限管理", Type: "menu", Path: "/system/permissions", Sort: 25},
	{Code: "system:rbac:identity", Name: "身份映射管理", Type: "menu", Path: "/system/identities", Sort: 26},
	{Code: "system:rbac:user", Name: "用户管理", Type: "menu", Path: "/system/users", Sort: 27},

	// MQ 数据源管理权限
	{Code: "data:mq-source:view", Name: "查看消息队列", Type: "menu", Path: "/data/mq-sources", Sort: 30},
	{Code: "data:mq-source:add", Name: "新增数据源", Type: "action", Method: "POST", Path: "/api/v1/mq-sources/add", Sort: 31},
	{Code: "data:mq-source:edit", Name: "编辑数据源", Type: "action", Method: "POST", Path: "/api/v1/mq-sources/:id/edit", Sort: 32},
	{Code: "data:mq-source:delete", Name: "删除数据源", Type: "action", Method: "POST", Path: "/api/v1/mq-sources/:id/delete", Sort: 33},
	{Code: "data:mq-source:test", Name: "测试连接", Type: "action", Method: "POST", Path: "/api/v1/mq-sources/:id/test", Sort: 34},

	// MQ 订阅管理权限
	{Code: "data:subscription:view", Name: "查看订阅管理", Type: "menu", Path: "/message/subscriptions", Sort: 40},
	{Code: "data:subscription:add", Name: "新增订阅", Type: "action", Method: "POST", Path: "/api/v1/subscriptions/add", Sort: 41},
	{Code: "data:subscription:edit", Name: "编辑订阅", Type: "action", Method: "POST", Path: "/api/v1/subscriptions/:id/edit", Sort: 42},
	{Code: "data:subscription:delete", Name: "删除订阅", Type: "action", Method: "POST", Path: "/api/v1/subscriptions/:id/delete", Sort: 43},
	{Code: "data:subscription:start", Name: "启动订阅", Type: "action", Method: "POST", Path: "/api/v1/subscriptions/:id/start", Sort: 44},
	{Code: "data:subscription:stop", Name: "停止订阅", Type: "action", Method: "POST", Path: "/api/v1/subscriptions/:id/stop", Sort: 45},

	// 消费日志权限
	{Code: "data:consume-log:view", Name: "查看消费日志", Type: "menu", Path: "/logs/consume-logs", Sort: 50},
}

func InitRbacSeedData() {
	operator := "system"

	err := ensureRbacAuthzModeSetting(operator)
	if err != nil {
		logrus.Errorf("初始化 RBAC 鉴权模式失败: %v", err)
		return
	}

	role, err := models.AddRoleIfNotExists(&models.RbacRole{
		IDModel: models.IDModel{
			CreatedBy:  operator,
			ModifiedBy: operator,
		},
		Code:        "super_admin",
		Name:        "超级管理员",
		Description: "系统超级管理员",
		Status:      1,
	})
	if err != nil {
		logrus.Errorf("初始化 super_admin 角色失败: %v", err)
		return
	}

	group, err := models.AddUserGroupIfNotExists(&models.RbacUserGroup{
		IDModel: models.IDModel{
			CreatedBy:  operator,
			ModifiedBy: operator,
		},
		Code:        "default_group",
		Name:        "默认用户组",
		Description: "系统默认用户组",
		Status:      1,
	})
	if err != nil {
		logrus.Errorf("初始化默认用户组失败: %v", err)
		return
	}

	err = models.AssignRoleToGroupIfNotExists(group.ID, role.ID, operator)
	if err != nil {
		logrus.Errorf("初始化默认用户组角色关系失败: %v", err)
		return
	}

	for _, item := range defaultPermissionSeeds {
		permission, addErr := models.AddPermissionIfNotExists(&models.RbacPermission{
			IDModel: models.IDModel{
				CreatedBy:  operator,
				ModifiedBy: operator,
			},
			Code:   item.Code,
			Name:   item.Name,
			Type:   item.Type,
			Method: item.Method,
			Path:   item.Path,
			Sort:   item.Sort,
			Status: 1,
		})
		if addErr != nil {
			logrus.Errorf("初始化权限 %s 失败: %v", item.Code, addErr)
			return
		}
		if relErr := models.AssignPermissionToRoleIfNotExists(role.ID, permission.ID, operator); relErr != nil {
			logrus.Errorf("初始化角色权限关系 %s 失败: %v", item.Code, relErr)
			return
		}
	}

	adminUser, err := models.GetUserByUsername("admin")
	if err != nil {
		logrus.Errorf("查询 admin 账号失败: %v", err)
		return
	}
	if err = models.AssignRoleToUserIfNotExists(adminUser.ID, role.ID, operator); err != nil {
		logrus.Errorf("绑定 admin 到 super_admin 失败: %v", err)
		return
	}
	if err = models.AssignUserToGroupIfNotExists(adminUser.ID, group.ID, operator); err != nil {
		logrus.Errorf("绑定 admin 到默认用户组失败: %v", err)
	}
}

func ensureRbacAuthzModeSetting(operator string) error {
	current, err := models.GetSettingByKey("rbac", "authz_mode")
	if err != nil {
		return err
	}
	if current.ID > 0 {
		return nil
	}
	return models.AddOneSetting(models.Settings{
		IDModel: models.IDModel{
			CreatedBy:  operator,
			ModifiedBy: operator,
		},
		Section: "rbac",
		Key:     "authz_mode",
		Value:   "monitor",
	})
}
