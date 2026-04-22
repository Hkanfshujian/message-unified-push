package casdoor_service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"ops-message-unified-push/models"
	"ops-message-unified-push/pkg/util"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// CasdoorUser Casdoor 返回的用户信息
type CasdoorUser struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
	Displayname string `json:"displayName"`
}

// TokenResponse Casdoor Token 响应
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`
	Scope        string `json:"scope"`
}

// UserInfoResponse Casdoor 用户信息响应
type UserInfoResponse struct {
	Status string      `json:"status"`
	Msg    string      `json:"msg"`
	Data   CasdoorUser `json:"data"`
}

// IDTokenStore id_token 存储（内存实现）
type IDTokenStore struct {
	mu     sync.RWMutex
	tokens map[uint]string // userID -> idToken
}

var (
	globalIDTokenStore *IDTokenStore
	idTokenStoreOnce   sync.Once
)

// GetIDTokenStore 获取全局 id_token 存储实例
func GetIDTokenStore() *IDTokenStore {
	idTokenStoreOnce.Do(func() {
		globalIDTokenStore = &IDTokenStore{
			tokens: make(map[uint]string),
		}
	})
	return globalIDTokenStore
}

// Store 存储用户的 id_token
func (s *IDTokenStore) Store(userID uint, idToken string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tokens[userID] = idToken
}

// Get 获取用户的 id_token
func (s *IDTokenStore) Get(userID uint) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	token, ok := s.tokens[userID]
	return token, ok
}

// Delete 删除用户的 id_token
func (s *IDTokenStore) Delete(userID uint) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tokens, userID)
}

// CasdoorService Casdoor 认证服务
type CasdoorService struct {
	Config *CasdoorConfig
}

// NewCasdoorService 创建 Casdoor 服务实例
func NewCasdoorService() (*CasdoorService, error) {
	config, err := LoadCasdoorConfig()
	if err != nil {
		return nil, fmt.Errorf("加载配置失败: %w", err)
	}
	if err := config.Validate(); err != nil {
		return nil, err
	}
	return &CasdoorService{Config: config}, nil
}

// BuildAuthURL 构建授权 URL
func (s *CasdoorService) BuildAuthURL(state, nonce string) string {
	params := url.Values{}
	params.Set("client_id", s.Config.ClientID)
	params.Set("redirect_uri", s.Config.RedirectURI)
	params.Set("response_type", "code")
	params.Set("scope", "openid profile email")
	params.Set("state", state)
	params.Set("nonce", nonce)
	return s.Config.GetAuthURL() + "?" + params.Encode()
}

// ExchangeToken 用授权码换取 Token
func (s *CasdoorService) ExchangeToken(ctx context.Context, code string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", s.Config.ClientID)
	data.Set("client_secret", s.Config.ClientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", s.Config.RedirectURI)

	req, err := http.NewRequestWithContext(ctx, "POST", s.Config.GetTokenURL(), strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Token 交换失败 [%d]: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("响应中无 access_token: %s", string(body))
	}

	return &tokenResp, nil
}

// GetUserInfo 获取用户信息
func (s *CasdoorService) GetUserInfo(ctx context.Context, accessToken string) (*CasdoorUser, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", s.Config.GetUserInfoURL(), nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取用户信息失败 [%d]: %s", resp.StatusCode, string(body))
	}

	var result UserInfoResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if result.Status != "ok" {
		return nil, fmt.Errorf("获取用户信息失败: %s", result.Msg)
	}

	return &result.Data, nil
}

// HandleCallback 处理回调，返回本地 token 和用户信息
func (s *CasdoorService) HandleCallback(ctx context.Context, code string) (localToken string, user *models.Auth, err error) {
	logrus.Info("[Casdoor] 开始处理回调")

	// 1. 用授权码换取 Token
	tokenResp, err := s.ExchangeToken(ctx, code)
	if err != nil {
		logrus.Errorf("[Casdoor] Token 交换失败: %v", err)
		return "", nil, fmt.Errorf("Token 交换失败: %w", err)
	}
	logrus.Info("[Casdoor] Token 交换成功")

	// 2. 获取用户信息
	casdoorUser, err := s.GetUserInfo(ctx, tokenResp.AccessToken)
	if err != nil {
		logrus.Errorf("[Casdoor] 获取用户信息失败: %v", err)
		return "", nil, fmt.Errorf("获取用户信息失败: %w", err)
	}
	logrus.Infof("[Casdoor] 用户信息获取成功: %s (%s)", casdoorUser.Name, casdoorUser.Email)

	// 3. 映射或创建本地用户
	user, err = s.mapOrCreateUser(casdoorUser)
	if err != nil {
		logrus.Errorf("[Casdoor] 用户映射失败: %v", err)
		return "", nil, fmt.Errorf("用户映射失败: %w", err)
	}
	logrus.Infof("[Casdoor] 用户映射成功: ID=%d, Username=%s", user.ID, user.Username)

	// 4. 生成本地 JWT Token
	localToken, err = util.GenerateToken(user.Username, "", 7) // 7天有效期
	if err != nil {
		logrus.Errorf("[Casdoor] 生成 Token 失败: %v", err)
		return "", nil, fmt.Errorf("生成 Token 失败: %w", err)
	}

	// 5. 存储 id_token（用于统一登出）
	if tokenResp.IDToken != "" {
		GetIDTokenStore().Store(uint(user.ID), tokenResp.IDToken)
		logrus.Info("[Casdoor] id_token 已存储")
	}

	logrus.Info("[Casdoor] 登录流程完成")
	return localToken, user, nil
}

// mapOrCreateUser 映射或创建本地用户
func (s *CasdoorService) mapOrCreateUser(casdoorUser *CasdoorUser) (*models.Auth, error) {
	// 优先用 Casdoor ID 查找
	casdoorSub := casdoorUser.ID
	if casdoorSub == "" {
		casdoorSub = casdoorUser.Name
	}

	// 查找已有绑定
	identity, err := models.GetAuthIdentityByProviderSub("casdoor", casdoorSub)
	if err == nil && identity != nil {
		// 已绑定，获取用户并确保渠道字段正确
		user, err := models.GetUserByID(identity.UserID)
		if err != nil {
			return nil, err
		}
		// 确保渠道字段为 casdoor
		if user.Channel != "casdoor" {
			models.UpdateAuthChannelInfo(user.ID, "casdoor", casdoorSub)
			user.Channel = "casdoor"
		}
		return user, nil
	}

	// 尝试用用户名查找
	username := sanitizeUsername(casdoorUser.Name)
	existingUser, err := models.GetUserByUsername(username)
	if err == nil && existingUser != nil {
		// 用户存在，创建绑定并更新渠道信息
		_ = models.AddAuthIdentity(&models.AuthIdentity{
			UserID:      existingUser.ID,
			Provider:    "casdoor",
			ExternalSub: casdoorSub,
		})
		// 更新用户渠道为 casdoor
		models.UpdateAuthChannelInfo(existingUser.ID, "casdoor", casdoorSub)
		existingUser.Channel = "casdoor"
		return existingUser, nil
	}

	// 用户不存在
	if !s.Config.AutoCreateUser {
		return nil, errors.New("未找到本地用户，且未启用自动创建")
	}

	// 自动创建用户（直接创建 casdoor 渠道用户）
	username, err = buildUniqueUsername(username)
	if err != nil {
		return nil, err
	}

	newUser, err := models.AddCasdoorUser(username, casdoorSub)
	if err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	// 创建身份绑定
	_ = models.AddAuthIdentity(&models.AuthIdentity{
		UserID:      newUser.ID,
		Provider:    "casdoor",
		ExternalSub: casdoorSub,
	})

	// 分配默认组
	if s.Config.DefaultGroupID > 0 {
		_ = models.AssignUserToGroupIfNotExists(newUser.ID, s.Config.DefaultGroupID, "system")
	}

	logrus.Infof("[Casdoor] 自动创建用户: %s (ID: %d)", username, newUser.ID)
	return newUser, nil
}

// BuildLogoutURL 构建统一登出 URL
func (s *CasdoorService) BuildLogoutURL(userID uint, redirectURI string) string {
	idToken, ok := GetIDTokenStore().Get(userID)
	if !ok {
		logrus.Warn("[Casdoor] 未找到 id_token，无法构建登出 URL")
		return ""
	}

	// 删除存储的 id_token
	GetIDTokenStore().Delete(userID)

	params := url.Values{}
	params.Set("id_token_hint", idToken)
	if redirectURI != "" {
		params.Set("post_logout_redirect_uri", redirectURI)
	}

	return s.Config.GetLogoutURL() + "?" + params.Encode()
}

// sanitizeUsername 清理用户名
func sanitizeUsername(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.ToLower(raw)
	// 替换特殊字符
	reg := regexp.MustCompile(`[^a-z0-9_]`)
	raw = reg.ReplaceAllString(raw, "_")
	// 去除连续下划线
	reg = regexp.MustCompile(`_+`)
	raw = reg.ReplaceAllString(raw, "_")
	raw = strings.Trim(raw, "_")
	if raw == "" {
		raw = "user"
	}
	return raw
}

// buildUniqueUsername 构建唯一用户名
func buildUniqueUsername(base string) (string, error) {
	candidate := base
	for i := 0; i < 20; i++ {
		_, err := models.GetUserByUsername(candidate)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return candidate, nil
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return "", err
		}
		candidate = fmt.Sprintf("%s_%d", base, time.Now().Unix()%100000+int64(i))
	}
	return "", errors.New("生成唯一用户名失败")
}
