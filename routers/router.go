package routers

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"ops-message-unified-push/middleware"
	"ops-message-unified-push/pkg/setting"
	"ops-message-unified-push/routers/api"
	v1 "ops-message-unified-push/routers/api/v1"
	v2 "ops-message-unified-push/routers/api/v2"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// AppendCors 添加是否跨域（所有模式开启）
func AppendCors(app *gin.Engine) {
	app.Use(middleware.Cors())
}

// AppendServerStaticHtmlWithPrefix 启用返回打包的静态文件（支持路径前缀）
func AppendServerStaticHtmlWithPrefix(router gin.IRouter, f embed.FS, pathPrefix string) {
	if setting.ServerSetting.EmbedHtml == "disable" {
		return
	}

	assets, _ := fs.Sub(f, "web/dist/assets")
	dist, _ := fs.Sub(f, "web/dist")

	// 根据是否有路径前缀来设置静态文件路由
	if pathPrefix != "" {
		// 有路径前缀时，使用相对路径
		if r, ok := router.(*gin.RouterGroup); ok {
			r.Use(middleware.StaticCacheMiddleware())
			r.StaticFS("/assets", http.FS(assets))
			r.GET("/", func(ctx *gin.Context) {
				// 读取 index.html
				indexFile, err := dist.Open("index.html")
				if err != nil {
					ctx.String(http.StatusInternalServerError, "Failed to load index.html")
					return
				}
				defer indexFile.Close()

				// 读取文件内容
				content, err := io.ReadAll(indexFile)
				if err != nil {
					ctx.String(http.StatusInternalServerError, "Failed to read index.html")
					return
				}

				htmlContent := string(content)

				// 注入 base 标签和配置脚本
				// base 标签必须在 head 的最前面，确保所有相对路径都基于这个 base
				baseTag := fmt.Sprintf(`<base href="%s/">`, pathPrefix)
				configScript := fmt.Sprintf(`<script>window.__URL_PATH_PREFIX__ = '%s';</script>`, pathPrefix)

				// 在 <head> 标签后立即插入 base 标签
				htmlContent = strings.Replace(htmlContent, "<head>", "<head>"+baseTag, 1)
				// 在 </head> 标签前注入配置
				htmlContent = strings.Replace(htmlContent, "</head>", configScript+"</head>", 1)

				ctx.Header("Content-Type", "text/html; charset=utf-8")
				ctx.String(http.StatusOK, htmlContent)
			})
		}
	} else {
		// 无路径前缀时，使用原有逻辑
		if r, ok := router.(*gin.Engine); ok {
			r.Use(middleware.StaticCacheMiddleware())
			r.StaticFS("assets/", http.FS(assets))
			r.GET("/", func(ctx *gin.Context) {
				ctx.FileFromFS("/", http.FS(dist))
			})
		}
	}
}

// AppendServerStaticHtml 启用返回打包的静态文件（保留向后兼容）
func AppendServerStaticHtml(app *gin.Engine, f embed.FS) {
	AppendServerStaticHtmlWithPrefix(app, f, "")
}

// InitRouter 初始化路由
func InitRouter(f embed.FS) *gin.Engine {
	app := gin.New()
	app.Use(middleware.LogMiddleware())
	app.Use(gin.Recovery())

	AppendCors(app)

	// 获取 URL 前缀
	pathPrefix := setting.ServerSetting.UrlPrefix
	if pathPrefix != "" && pathPrefix[0] != '/' {
		pathPrefix = "/" + pathPrefix
	}

	// 如果有路径前缀，创建路由组
	var router gin.IRouter
	if pathPrefix != "" {
		router = app.Group(pathPrefix)
	} else {
		router = app
	}

	AppendServerStaticHtmlWithPrefix(router, f, pathPrefix)
	router.Static("/uploads", "./data/uploads")
	router.Static("/storage", "./data")
	router.Static("/public/uploads", "./data/uploads")
	router.Static("/public/storage/local", "./data")

	// 健康检查接口（无需认证）
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"time":   time.Now().Unix(),
		})
	})

	router.POST("/auth", api.GetAuth)
	router.POST("/auth/register", api.RegisterAuth)
	router.GET("/auth/public-config", api.GetPublicAuthConfig)
	
	// Casdoor 登录
	router.GET("/auth/casdoor/login", api.CasdoorLogin)
	router.GET("/auth/casdoor/callback", api.CasdoorCallback)
	router.GET("/auth/casdoor/logout/callback", api.CasdoorLogoutCallback)
	router.GET("/auth/casdoor/status", api.CasdoorStatus)
	apiV1 := router.Group("/api/v1")
	apiV1.Use(middleware.JWT())
	{
		// sendways
		apiV1.POST("/sendways/add", middleware.RequirePermission("message:sendways:add"), v1.AddMsgSendWay)
		apiV1.POST("/sendways/delete", middleware.RequirePermission("message:sendways:delete"), v1.DeleteMsgSendWay)
		apiV1.POST("/sendways/edit", middleware.RequirePermission("message:sendways:edit"), v1.EditSendWay)
		apiV1.POST("/sendways/test", middleware.RequirePermission("message:sendways:test"), v1.TestSendWay)
		apiV1.GET("/sendways/list", middleware.RequirePermission("message:sendways:view"), v1.GetMsgSendWayList)
		apiV1.GET("/sendways/get", middleware.RequirePermission("message:sendways:view"), v1.GetMsgSendWay)

		apiV1.GET("/sendlogs/list", middleware.RequirePermission("message:sendlogs:view"), v1.GetTaskSendLogsList)

		// settings
		apiV1.POST("/settings/setpasswd", middleware.RequirePermission("system:settings:edit"), v1.EditPasswd)
		apiV1.POST("/profile/password", middleware.RequirePermission("profile:settings:edit"), v1.EditPasswd)
		apiV1.POST("/auth/casdoor/logout", api.CasdoorLogout) // Casdoor 统一登出
		apiV1.POST("/settings/set", middleware.RequirePermission("system:settings:edit"), v1.EditSettings)
		apiV1.POST("/settings/reset", middleware.RequirePermission("system:settings:edit"), v1.RestDefaultSettings)
		apiV1.GET("/settings/getsetting", v1.GetUserSetting)
		apiV1.GET("/system/auth-config", middleware.RequirePermission("system:settings:view"), v1.GetSystemAuthConfig)
		apiV1.POST("/system/auth-config", middleware.RequirePermission("system:settings:edit"), v1.UpdateSystemAuthConfig)
		apiV1.POST("/system/site-logo/upload", middleware.RequirePermission("system:settings:edit"), v1.UploadSiteLogo)
		apiV1.POST("/system/site-logo/clear", middleware.RequirePermission("system:settings:edit"), v1.ClearSiteLogo)
		apiV1.GET("/system/storage-config", middleware.RequirePermission("system:settings:view"), v1.GetSystemStorageConfig)
		apiV1.GET("/system/storage-config/local-directories", middleware.RequirePermission("system:settings:view"), v1.GetSystemStorageLocalDirectories)
		apiV1.GET("/system/storage-config/local-files", middleware.RequirePermission("system:settings:view"), v1.GetSystemStorageLocalFiles)
		apiV1.GET("/system/storage-config/s3-objects", middleware.RequirePermission("system:settings:view"), v1.GetSystemStorageS3Objects)
		apiV1.POST("/system/storage-config/local-directories", middleware.RequirePermission("system:settings:edit"), v1.CreateSystemStorageLocalDirectory)
		apiV1.POST("/system/storage-config/delete-file", middleware.RequirePermission("system:settings:edit"), v1.DeleteSystemStorageFile)
		apiV1.POST("/system/storage-config/upload-file", middleware.RequirePermission("system:settings:edit"), v1.UploadSystemStorageFile)
		apiV1.POST("/system/storage-config/test-local-upload", middleware.RequirePermission("system:settings:edit"), v1.UploadSystemStorageFile)
		apiV1.POST("/system/storage-config", middleware.RequirePermission("system:settings:edit"), v1.UpdateSystemStorageConfig)
		apiV1.POST("/system/storage-config/test", middleware.RequirePermission("system:settings:edit"), v1.TestSystemStorageConfig)
		apiV1.GET("/profile/theme", middleware.RequirePermission("profile:settings:view"), v1.GetProfileTheme)
		apiV1.POST("/profile/theme", middleware.RequirePermission("profile:settings:edit"), v1.UpdateProfileTheme)

		// login logs
		apiV1.GET("/loginlogs/recent", middleware.RequirePermission("system:loginlogs:view"), v1.GetRecentLoginLogs)

		// statistic
		apiV1.GET("/statistic", middleware.RequirePermission("dashboard:view"), v1.GetStatisticData)
		apiV1.GET("/statistic/task", middleware.RequirePermission("dashboard:view"), v1.GetSendStatsByTask)

		// cronMessage
		apiV1.POST("/cronmessages/addone", middleware.RequirePermission("message:cron:add"), v1.AddCronMsgTask)
		apiV1.GET("/cronmessages/list", middleware.RequirePermission("message:cron:view"), v1.GetCronMsgList)
		apiV1.POST("/cronmessages/delete", middleware.RequirePermission("message:cron:delete"), v1.DeleteCronMsgTask)
		apiV1.POST("/cronmessages/edit", middleware.RequirePermission("message:cron:edit"), v1.EditCronMsgTask)
		apiV1.POST("/cronmessages/sendnow", middleware.RequirePermission("message:cron:sendnow"), v1.SendNowCronMsg)

		// messageTemplate
		apiV1.GET("/templates/list", middleware.RequirePermission("message:template:view"), v1.GetMessageTemplateList)
		apiV1.GET("/templates/get", middleware.RequirePermission("message:template:view"), v1.GetMessageTemplate)
		apiV1.POST("/templates/add", middleware.RequirePermission("message:template:add"), v1.AddMessageTemplate)
		apiV1.POST("/templates/edit", middleware.RequirePermission("message:template:edit"), v1.EditMessageTemplate)
		apiV1.POST("/templates/delete", middleware.RequirePermission("message:template:delete"), v1.DeleteMessageTemplate)
		apiV1.POST("/templates/preview", middleware.RequirePermission("message:template:preview"), v1.PreviewMessageTemplate)

		// messageTemplate instances
		apiV1.GET("/templates/ins/get", middleware.RequireAnyPermission("message:template:view", "message:template:instance"), v1.GetTemplateWithIns)
		apiV1.POST("/templates/ins/addone", middleware.RequirePermission("message:template:instance"), v1.AddTemplateIns)
		apiV1.POST("/templates/ins/delete", middleware.RequirePermission("message:template:instance"), v1.DeleteTemplateIns)
		apiV1.POST("/templates/ins/update_enable", middleware.RequirePermission("message:template:instance"), v1.UpdateTemplateInsEnable)
		apiV1.POST("/templates/ins/update_config", middleware.RequirePermission("message:template:instance"), v1.UpdateTemplateInsConfig)
		apiV1.GET("/templates/relations", middleware.RequirePermission("message:template:view"), v1.GetTemplateRelations)

		apiV1.GET("/rbac/me/permissions", v1.GetCurrentUserPermissions)
		apiV1.GET("/rbac/roles", middleware.RequireAnyPermission("system:rbac:role", "system:rbac:group"), v1.GetRbacRoles)
		apiV1.POST("/rbac/roles", middleware.RequirePermission("system:rbac:role"), v1.AddRbacRole)
		apiV1.POST("/rbac/roles/edit", middleware.RequirePermission("system:rbac:role"), v1.EditRbacRole)
		apiV1.POST("/rbac/roles/delete", middleware.RequirePermission("system:rbac:role"), v1.DeleteRbacRole)
		apiV1.GET("/rbac/roles/permissions", middleware.RequirePermission("system:rbac:role"), v1.GetRolePermissionIDs)
		apiV1.POST("/rbac/roles/assign-permissions", middleware.RequirePermission("system:rbac:role"), v1.AssignPermissionsToRole)

		apiV1.GET("/rbac/groups", middleware.RequireAnyPermission("system:rbac:group", "system:rbac:role"), v1.GetRbacGroups)
		apiV1.POST("/rbac/groups", middleware.RequirePermission("system:rbac:group"), v1.AddRbacGroup)
		apiV1.POST("/rbac/groups/edit", middleware.RequirePermission("system:rbac:group"), v1.EditRbacGroup)
		apiV1.POST("/rbac/groups/delete", middleware.RequirePermission("system:rbac:group"), v1.DeleteRbacGroup)
		apiV1.GET("/rbac/groups/roles", middleware.RequirePermission("system:rbac:group"), v1.GetGroupRoleIDs)
		apiV1.GET("/rbac/groups/members", middleware.RequirePermission("system:rbac:group"), v1.GetGroupMemberIDs)
		apiV1.POST("/rbac/groups/assign-roles", middleware.RequirePermission("system:rbac:group"), v1.AssignRolesToGroup)
		apiV1.POST("/rbac/groups/assign-members", middleware.RequirePermission("system:rbac:group"), v1.AssignMembersToGroup)

		apiV1.GET("/rbac/permissions", middleware.RequireAnyPermission("system:rbac:permission", "system:rbac:role"), v1.GetRbacPermissions)
		apiV1.POST("/rbac/permissions", middleware.RequirePermission("system:rbac:permission"), v1.AddRbacPermission)
		apiV1.POST("/rbac/permissions/edit", middleware.RequirePermission("system:rbac:permission"), v1.EditRbacPermission)

		apiV1.GET("/rbac/users", middleware.RequireAnyPermission("system:rbac:role", "system:rbac:group"), v1.GetRbacUsers)
		apiV1.GET("/rbac/users/manage", middleware.RequirePermission("system:rbac:user"), v1.GetManageUsers)
		apiV1.POST("/rbac/users/manage", middleware.RequirePermission("system:rbac:user"), v1.AddManageUser)
		apiV1.POST("/rbac/users/manage/edit", middleware.RequirePermission("system:rbac:user"), v1.EditManageUser)
		apiV1.POST("/rbac/users/manage/delete", middleware.RequirePermission("system:rbac:user"), v1.DeleteManageUser)
		apiV1.GET("/rbac/users/role-ids", middleware.RequirePermission("system:rbac:role"), v1.GetUserRoleIDs)
		apiV1.GET("/rbac/users/group-ids", middleware.RequirePermission("system:rbac:group"), v1.GetUserGroupIDs)
		apiV1.POST("/rbac/users/assign-roles", middleware.RequirePermission("system:rbac:role"), v1.AssignRolesToUser)
		apiV1.POST("/rbac/users/assign-groups", middleware.RequirePermission("system:rbac:group"), v1.AssignGroupsToUser)

		// MQ 数据源管理
		mqSourceCtrl := v1.MQSourceController{}
		apiV1.GET("/mq-sources/list", middleware.RequirePermission("data:mq-source:view"), mqSourceCtrl.GetMQSourceList)
		apiV1.GET("/mq-sources/:id", middleware.RequirePermission("data:mq-source:view"), mqSourceCtrl.GetMQSourceByID)
		apiV1.POST("/mq-sources/add", middleware.RequirePermission("data:mq-source:add"), mqSourceCtrl.AddMQSource)
		apiV1.POST("/mq-sources/:id/edit", middleware.RequirePermission("data:mq-source:edit"), mqSourceCtrl.EditMQSource)
		apiV1.POST("/mq-sources/:id/delete", middleware.RequirePermission("data:mq-source:delete"), mqSourceCtrl.DeleteMQSource)
		apiV1.POST("/mq-sources/:id/test", middleware.RequirePermission("data:mq-source:test"), mqSourceCtrl.TestMQSource)
		apiV1.POST("/mq-sources/test-config", middleware.RequirePermission("data:mq-source:test"), mqSourceCtrl.TestMQSourceConfig)

		// MQ 订阅管理
		subscriptionCtrl := v1.SubscriptionController{}
		apiV1.GET("/subscriptions/list", middleware.RequirePermission("data:subscription:view"), subscriptionCtrl.GetSubscriptionList)
		apiV1.GET("/subscriptions/:id", middleware.RequirePermission("data:subscription:view"), subscriptionCtrl.GetSubscriptionByID)
		apiV1.POST("/subscriptions/add", middleware.RequirePermission("data:subscription:add"), subscriptionCtrl.AddSubscription)
		apiV1.POST("/subscriptions/regex-test", middleware.RequireAnyPermission("data:subscription:add", "data:subscription:edit"), subscriptionCtrl.TestSubscriptionRegex)
		apiV1.POST("/subscriptions/:id/edit", middleware.RequirePermission("data:subscription:edit"), subscriptionCtrl.EditSubscription)
		apiV1.POST("/subscriptions/:id/delete", middleware.RequirePermission("data:subscription:delete"), subscriptionCtrl.DeleteSubscription)
		apiV1.POST("/subscriptions/:id/start", middleware.RequirePermission("data:subscription:start"), subscriptionCtrl.StartSubscription)
		apiV1.POST("/subscriptions/:id/stop", middleware.RequirePermission("data:subscription:stop"), subscriptionCtrl.StopSubscription)

		// 消费日志
		consumeLogCtrl := v1.ConsumeLogController{}
		apiV1.GET("/consume-logs/list", middleware.RequirePermission("data:consume-log:view"), consumeLogCtrl.GetConsumeLogList)
		apiV1.GET("/consume-logs/:id", middleware.RequirePermission("data:consume-log:view"), consumeLogCtrl.GetConsumeLogByID)
		apiV1.GET("/consume-logs/stats", middleware.RequirePermission("data:consume-log:view"), consumeLogCtrl.GetConsumeStats)
	}

	// API v2
	apiV2 := router.Group("/api/v2")
	apiV2.Use(middleware.JWT())
	{
		// message/send - 使用模板发送消息
		apiV2.POST("/message/send", v2.DoSendMessageByTemplate)
	}

	return app
}
