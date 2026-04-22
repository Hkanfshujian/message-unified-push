package settings_service

import (
	"errors"
	"fmt"
	"ops-message-unified-push/models"
	"ops-message-unified-push/pkg/app"
	"ops-message-unified-push/pkg/constant"
	"ops-message-unified-push/pkg/util"
	"ops-message-unified-push/service/cron_service"
	"strconv"
	"strings"

	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

type UserSettings struct {
	UserName    string
	OldPassword string
	NewPassword string
}

// EditUserPasswd 用户设置密码
func (us *UserSettings) EditUserPasswd() error {
	ok, _ := models.CheckAuth(us.UserName, us.OldPassword)
	if !ok {
		return errors.New("旧密码校验失败！")
	}
	var user = map[string]interface{}{
		"password": util.EncodeMD5(us.NewPassword),
	}
	return models.EditUser(us.UserName, user)
}

// GetUserSetting 获取用户设置
func (us *UserSettings) GetUserSetting(section string) (map[string]string, error) {
	// 如果是site_config，优先从缓存获取
	if section == constant.SiteSettingSectionName {
		if IsSiteConfigCacheValid() {
			return GetSiteConfigCache(), nil
		}
	}

	result := make(map[string]string)
	settings, err := models.GetSettingBySection(section)
	if err != nil {
		return result, err
	}
	for _, setting := range settings {
		result[setting.Key] = setting.Value
	}

	// 如果是site_config，更新缓存
	if section == constant.SiteSettingSectionName {
		SetSiteConfigCache(result)
	}

	// 版本信息单独获取
	if section == constant.AboutSectionName {
		result = constant.LatestVersion
		// 添加内存使用信息
		memoryInfo := util.GetMemoryUsage()
		for key, value := range memoryInfo {
			result[key] = value
		}
	}
	return result, nil
}

// EditSettings 便捷自定义设置
func (us *UserSettings) EditSettings(section string, key string, value string, currentUser string) error {
	setting, _ := models.GetSettingByKey(section, key)
	if setting.ID <= 0 {
		err := models.AddOneSetting(models.Settings{
			IDModel: models.IDModel{
				CreatedBy:  currentUser,
				ModifiedBy: currentUser,
			},
			Section: section,
			Key:     key,
			Value:   value,
		})
		// 如果是site_config，清除缓存
		if section == constant.SiteSettingSectionName {
			ClearSiteConfigCache()
		}
		if section == constant.MQStatusPolicySectionName && key == constant.MQStatusPolicyLogLevelKeyName {
			applyRuntimeLogLevel(value)
		}
		return err
	} else {
		if value == "" {
			return nil
		}
		oldValue := setting.Value
		data := make(map[string]interface{})
		data["section"] = section
		data["key"] = key
		data["value"] = value
		data["modified_by"] = currentUser
		err := models.EditSetting(setting.ID, data)
		// 更新日志清理配置
		if isLogCleanupSection(section) {
			if key == constant.LogsCleanCronKeyName || key == constant.LogsCleanEnabledKeyName {
				cronService := cron_service.CronService{}
				// 获取当前 section 下最新的 cron 和 enabled 值
				cronSetting, _ := models.GetSettingByKey(section, constant.LogsCleanCronKeyName)
				enabledSetting, _ := models.GetSettingByKey(section, constant.LogsCleanEnabledKeyName)
				enabled := enabledSetting.Value == "true"

				switch section {
				case constant.LogsCleanSectionName:
					cronService.UpdateLogsCronRun(cronSetting.Value, enabled)
				case constant.ConsumeLogsCleanSectionName:
					cronService.UpdateConsumeLogsCronRun(cronSetting.Value, enabled)
				case constant.LoginLogsCleanSectionName:
					cronService.UpdateLoginLogsCronRun(cronSetting.Value, enabled)
				}
			}
		}
		// 更新消息队列状态自动更新策略
		if section == constant.MQStatusPolicySectionName {
			if key == constant.MQStatusPolicyEnabledKeyName || key == constant.MQStatusPolicyIntervalSecondsKeyName {
				cronService := cron_service.CronService{}
				enabledSetting, _ := models.GetSettingByKey(constant.MQStatusPolicySectionName, constant.MQStatusPolicyEnabledKeyName)
				intervalSetting, _ := models.GetSettingByKey(constant.MQStatusPolicySectionName, constant.MQStatusPolicyIntervalSecondsKeyName)
				intervalSeconds, _ := strconv.Atoi(intervalSetting.Value)
				if intervalSeconds <= 0 {
					intervalSeconds = 300
				}
				cronService.UpdateMQStatusPolicyRun(enabledSetting.Value == "true", intervalSeconds)
			}
			if key == constant.MQStatusPolicyLogLevelKeyName {
				applyRuntimeLogLevel(value)
				addSystemSettingAuditLog(currentUser, section, key, oldValue, value)
			}
		}
		// 如果是site_config，清除缓存
		if section == constant.SiteSettingSectionName {
			ClearSiteConfigCache()
		}
		return err
	}
}

// 站点自定义的结构
type SiteConfig struct {
	Title                string `json:"title" validate:"omitempty,min=1,max=50" label:"网站标题"`
	Slogan               string `json:"slogan" validate:"omitempty,min=1,max=50" label:"网站slogan"`
	Logo                 string `json:"logo" validate:"omitempty,min=1" label:"logo"`
	LogoStorageProfileID string `json:"logo_storage_profile_id" validate:"omitempty,max=20" label:"logo存储ID"`
	CookieExpDays        string `json:"cookie_exp_days" validate:"omitempty,numeric,min=1,max=365" label:"cookie过期天数"`
	ThemeColor           string `json:"theme_color" validate:"omitempty,max=20" label:"主题颜色"`
	SloganInitialEnabled string `json:"slogan_initial_enabled" validate:"omitempty,oneof=true false" label:"标语首字母开关"`
	ChannelTestMessage   string `json:"channel_test_message" validate:"omitempty,max=2000" label:"渠道测试默认文案"`
}

type LogConfig struct {
	Cron    string `json:"cron" validate:"required,cron" label:"日志定时表达式"`
	KeepNum string `json:"keep_num" validate:"required,min=1,max=50" label:"日志保留数"`
	Enabled string `json:"enabled" validate:"required,oneof=true false" label:"是否启用"`
}

type AuthConfig struct {
	RegisterEnabled       string `json:"register_enabled" validate:"required,oneof=true false" label:"注册开关"`
	LocalDefaultGroupCode string `json:"local_default_group_code" validate:"omitempty,max=100" label:"本地默认用户组"`

	CasdoorEnabled          string `json:"casdoor_enabled" validate:"required,oneof=true false" label:"Casdoor启用状态"`
	CasdoorEndpoint         string `json:"casdoor_endpoint" validate:"omitempty,max=255" label:"Casdoor服务地址"`
	CasdoorClientID         string `json:"casdoor_client_id" validate:"omitempty,max=255" label:"Casdoor ClientID"`
	CasdoorClientSecret     string `json:"casdoor_client_secret" validate:"omitempty,max=255" label:"Casdoor ClientSecret"`
	CasdoorRedirectURI      string `json:"casdoor_redirect_uri" validate:"omitempty,max=255" label:"Casdoor回调地址"`
	CasdoorAuthPath         string `json:"casdoor_auth_path" validate:"omitempty,max=255" label:"Casdoor授权地址路径"`
	CasdoorTokenPath        string `json:"casdoor_token_path" validate:"omitempty,max=255" label:"Casdoor令牌地址路径"`
	CasdoorUserInfoPath     string `json:"casdoor_userinfo_path" validate:"omitempty,max=255" label:"Casdoor用户信息地址路径"`
	CasdoorLogoutPath       string `json:"casdoor_logout_path" validate:"omitempty,max=255" label:"Casdoor登出地址路径"`
	CasdoorAutoCreateUser   string `json:"casdoor_auto_create_user" validate:"required,oneof=true false" label:"Casdoor自动建档"`
	CasdoorDefaultGroupCode string `json:"casdoor_default_group_code" validate:"omitempty,max=100" label:"Casdoor默认用户组"`
	CasdoorButtonText       string `json:"casdoor_button_text" validate:"omitempty,max=50" label:"Casdoor登录按钮文案"`
	CasdoorButtonIcon       string `json:"casdoor_button_icon" validate:"omitempty,max=500" label:"Casdoor登录按钮图标"`
}

type StorageConfig struct {
	StorageProvider   string `json:"storage_provider" validate:"required,oneof=local s3" label:"存储驱动"`
	S3Endpoint        string `json:"s3_endpoint" validate:"omitempty,max=255" label:"S3 Endpoint"`
	S3Region          string `json:"s3_region" validate:"omitempty,max=100" label:"S3 Region"`
	S3Bucket          string `json:"s3_bucket" validate:"omitempty,max=100" label:"S3 Bucket"`
	S3AccessKey       string `json:"s3_access_key" validate:"omitempty,max=255" label:"S3 AccessKey"`
	S3SecretKey       string `json:"s3_secret_key" validate:"omitempty,max=255" label:"S3 SecretKey"`
	S3UseSSL          string `json:"s3_use_ssl" validate:"required,oneof=true false" label:"S3 SSL"`
	S3PublicBaseURL   string `json:"s3_public_base_url" validate:"omitempty,max=255" label:"S3 公网基础地址"`
	S3ProxyPublicRead string `json:"s3_proxy_public_read" validate:"required,oneof=true false" label:"S3 代理公开读取"`
	S3ObjectKeyPrefix string `json:"s3_object_key_prefix" validate:"omitempty,max=100" label:"S3 对象前缀"`
}

type MQStatusPolicyConfig struct {
	Enabled         string `json:"enabled" validate:"required,oneof=true false" label:"自动更新开关"`
	IntervalSeconds string `json:"interval_seconds" validate:"required,numeric" label:"自动更新频率(秒)"`
	LogLevel        string `json:"log_level" validate:"required,oneof=debug info warn error" label:"日志级别"`
}

// GetCookieExpDays 获取 cookie 过期天数，若无配置则返回默认值 1
func GetCookieExpDays() int {
	// 优先从缓存获取
	if IsSiteConfigCacheValid() {
		cache := GetSiteConfigCache()
		if expDays, ok := cache["cookie_exp_days"]; ok && expDays != "" {
			if days, err := strconv.Atoi(expDays); err == nil && days > 0 {
				return days
			}
		}
	}

	// 从数据库获取
	setting, _ := models.GetSettingByKey(constant.SiteSettingSectionName, "cookie_exp_days")
	if setting.ID > 0 && setting.Value != "" {
		if days, err := strconv.Atoi(setting.Value); err == nil && days > 0 {
			return days
		}
	}

	// 返回默认值
	return 1
}

// ValidateDiffSetting 校验不同的设置
func (us *UserSettings) ValidateDiffSetting(section string, data map[string]string) string {
	if section == constant.SiteSettingSectionName {
		var config SiteConfig
		config.Title = data["title"]
		config.Slogan = data["slogan"]
		config.Logo = data["logo"]
		config.LogoStorageProfileID = data["logo_storage_profile_id"]
		config.CookieExpDays = data["cookie_exp_days"]
		config.ThemeColor = data["theme_color"]
		config.SloganInitialEnabled = data["slogan_initial_enabled"]
		config.ChannelTestMessage = data["channel_test_message"]
		_, errStr := app.CommonPlaygroundValid(config)
		return errStr
	}
	if isLogCleanupSection(section) {
		var config LogConfig
		config.Cron = data["cron"]
		config.KeepNum = data["keep_num"]
		config.Enabled = data["enabled"]
		_, err := cron.ParseStandard(config.Cron)
		if err != nil {
			return fmt.Sprintf("%s 不是合法的cron表达式", config.Cron)
		}
		_, errStr := app.CommonPlaygroundValid(config)
		return errStr
	}
	if section == constant.AuthConfigSectionName {
		var config AuthConfig
		config.RegisterEnabled = data["register_enabled"]
		config.LocalDefaultGroupCode = data["local_default_group_code"]
		config.CasdoorEnabled = data["casdoor_enabled"]
		config.CasdoorEndpoint = data["casdoor_endpoint"]
		config.CasdoorClientID = data["casdoor_client_id"]
		config.CasdoorClientSecret = data["casdoor_client_secret"]
		config.CasdoorRedirectURI = data["casdoor_redirect_uri"]
		config.CasdoorAuthPath = data["casdoor_auth_path"]
		config.CasdoorTokenPath = data["casdoor_token_path"]
		config.CasdoorUserInfoPath = data["casdoor_userinfo_path"]
		config.CasdoorLogoutPath = data["casdoor_logout_path"]
		config.CasdoorAutoCreateUser = data["casdoor_auto_create_user"]
		config.CasdoorDefaultGroupCode = data["casdoor_default_group_code"]
		config.CasdoorButtonText = data["casdoor_button_text"]
		config.CasdoorButtonIcon = data["casdoor_button_icon"]
		_, errStr := app.CommonPlaygroundValid(config)
		return errStr
	}
	if section == constant.StorageConfigSectionName {
		var config StorageConfig
		config.StorageProvider = data["storage_provider"]
		config.S3Endpoint = data["s3_endpoint"]
		config.S3Region = data["s3_region"]
		config.S3Bucket = data["s3_bucket"]
		config.S3AccessKey = data["s3_access_key"]
		config.S3SecretKey = data["s3_secret_key"]
		config.S3UseSSL = data["s3_use_ssl"]
		config.S3PublicBaseURL = data["s3_public_base_url"]
		config.S3ProxyPublicRead = data["s3_proxy_public_read"]
		config.S3ObjectKeyPrefix = data["s3_object_key_prefix"]
		_, errStr := app.CommonPlaygroundValid(config)
		return errStr
	}
	if section == constant.MQStatusPolicySectionName {
		var config MQStatusPolicyConfig
		config.Enabled = data["enabled"]
		config.IntervalSeconds = data["interval_seconds"]
		config.LogLevel = strings.ToLower(strings.TrimSpace(data["log_level"]))
		if config.LogLevel == "" {
			config.LogLevel = "info"
		}
		_, errStr := app.CommonPlaygroundValid(config)
		if errStr != "" {
			return errStr
		}
		seconds, err := strconv.Atoi(config.IntervalSeconds)
		if err != nil {
			return "自动更新频率(秒)必须是整数"
		}
		if seconds < 10 || seconds > 86400 {
			return "自动更新频率(秒)范围必须在 10 ~ 86400 之间"
		}
		return errStr
	}
	return fmt.Sprintf("未知的section：%s", section)
}

func isLogCleanupSection(section string) bool {
	return section == constant.LogsCleanSectionName ||
		section == constant.ConsumeLogsCleanSectionName ||
		section == constant.LoginLogsCleanSectionName
}

func applyRuntimeLogLevel(level string) {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}
}

func addSystemSettingAuditLog(user, section, key, oldValue, newValue string) {
	if strings.TrimSpace(oldValue) == strings.TrimSpace(newValue) {
		return
	}
	status := 1
	logRecord := models.SendTasksLogs{
		TaskID:   "SYSCFGLOG001",
		Type:     "system_setting",
		Name:     "系统设置变更",
		Log:      fmt.Sprintf("用户[%s] 修改设置 %s.%s: '%s' -> '%s'", user, section, key, oldValue, newValue),
		Status:   &status,
		CallerIp: fmt.Sprintf("[SettingsAudit] %s", user),
	}
	if err := logRecord.Add(); err != nil {
		logrus.Errorf("写入系统设置审计日志失败: %v", err)
	}
}
