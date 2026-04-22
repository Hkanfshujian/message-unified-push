package casdoor_service

import (
	"errors"
	"ops-message-unified-push/models"
	"strings"
)

// CasdoorConfig Casdoor 配置
type CasdoorConfig struct {
	Enabled        bool
	Endpoint       string // Casdoor 服务地址
	ClientID       string
	ClientSecret   string
	RedirectURI    string // 回调地址
	AuthPath       string // 授权路径，默认 /login/oauth/authorize
	TokenPath      string // Token 路径，默认 /api/login/oauth/access_token
	UserInfoPath   string // 用户信息路径，默认 /api/get-account
	LogoutPath     string // 登出路径，默认 /api/logout
	AutoCreateUser bool   // 是否自动创建用户
	DefaultGroupID uint   // 新用户默认组
	ButtonText     string // 登录按钮文案
}

// getSettingValue 获取配置值
func getSettingValue(section, key string) string {
	setting, err := models.GetSettingByKey(section, key)
	if err != nil {
		return ""
	}
	return setting.Value
}

// LoadCasdoorConfig 从数据库加载 Casdoor 配置
func LoadCasdoorConfig() (*CasdoorConfig, error) {
	config := &CasdoorConfig{
		AuthPath:       "/login/oauth/authorize",
		TokenPath:      "/api/login/oauth/access_token",
		UserInfoPath:   "/api/get-account",
		LogoutPath:     "/api/logout",
		AutoCreateUser: true,
		DefaultGroupID: 2, // 默认普通用户组
	}

	// 读取 Casdoor 独立配置
	if val := getSettingValue("auth_config", "casdoor_enabled"); val != "" {
		config.Enabled = strings.EqualFold(val, "true") || val == "1"
	}
	if val := getSettingValue("auth_config", "casdoor_endpoint"); val != "" {
		config.Endpoint = strings.TrimRight(strings.TrimSpace(val), "/")
	}
	if val := getSettingValue("auth_config", "casdoor_client_id"); val != "" {
		config.ClientID = strings.TrimSpace(val)
	}
	if val := getSettingValue("auth_config", "casdoor_client_secret"); val != "" {
		config.ClientSecret = strings.TrimSpace(val)
	}
	if val := getSettingValue("auth_config", "casdoor_redirect_uri"); val != "" {
		config.RedirectURI = strings.TrimSpace(val)
	}
	if val := getSettingValue("auth_config", "casdoor_auto_create_user"); val != "" {
		config.AutoCreateUser = strings.EqualFold(val, "true") || val == "1"
	}
	if val := getSettingValue("auth_config", "casdoor_button_text"); val != "" {
		config.ButtonText = strings.TrimSpace(val)
	}

	// 自定义端点（如果配置了）
	if val := getSettingValue("auth_config", "casdoor_auth_path"); val != "" {
		config.AuthPath = strings.TrimSpace(val)
	}
	if val := getSettingValue("auth_config", "casdoor_token_path"); val != "" {
		config.TokenPath = strings.TrimSpace(val)
	}
	if val := getSettingValue("auth_config", "casdoor_userinfo_path"); val != "" {
		config.UserInfoPath = strings.TrimSpace(val)
	}
	if val := getSettingValue("auth_config", "casdoor_logout_path"); val != "" {
		config.LogoutPath = strings.TrimSpace(val)
	}

	// 默认组 ID
	if val := getSettingValue("auth_config", "casdoor_default_group_code"); val != "" {
		group, err := models.GetUserGroupByCode(val)
		if err == nil && group != nil {
			config.DefaultGroupID = group.ID
		}
	}

	return config, nil
}

// Validate 验证配置是否完整
func (c *CasdoorConfig) Validate() error {
	if !c.Enabled {
		return errors.New("Casdoor 登录未启用")
	}
	if c.Endpoint == "" {
		return errors.New("Casdoor 服务地址未配置")
	}
	if c.ClientID == "" {
		return errors.New("Client ID 未配置")
	}
	if c.ClientSecret == "" {
		return errors.New("Client Secret 未配置")
	}
	if c.RedirectURI == "" {
		return errors.New("回调地址未配置")
	}
	return nil
}

// GetAuthURL 获取完整的授权 URL
func (c *CasdoorConfig) GetAuthURL() string {
	return c.Endpoint + c.AuthPath
}

// GetTokenURL 获取完整的 Token URL
func (c *CasdoorConfig) GetTokenURL() string {
	return c.Endpoint + c.TokenPath
}

// GetUserInfoURL 获取完整的用户信息 URL
func (c *CasdoorConfig) GetUserInfoURL() string {
	return c.Endpoint + c.UserInfoPath
}

// GetLogoutURL 获取完整的登出 URL
func (c *CasdoorConfig) GetLogoutURL() string {
	return c.Endpoint + c.LogoutPath
}
