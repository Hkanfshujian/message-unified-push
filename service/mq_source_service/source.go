package mq_source_service

import (
	"fmt"
	"ops-message-unified-push/models"
	"net"
	"net/url"
	"strings"
	"time"
)

// MQSourceService 消息队列数据源服务
type MQSourceService struct {
	Name string
	Type string
}

type AddMQSourceRequest struct {
	Name        string `json:"name" binding:"required,max=200"`
	Type        string `json:"type" binding:"required,oneof=rocketmq kafka rabbitmq"`
	NamesrvAddr string `json:"namesrv_addr" binding:"required,max=500"`
	AccessKey   string `json:"access_key" binding:"max=200"`
	SecretKey   string `json:"secret_key" binding:"max=200"`
}

type EditMQSourceRequest struct {
	Name        string `json:"name" binding:"required,max=200"`
	Type        string `json:"type" binding:"required,oneof=rocketmq kafka rabbitmq"`
	NamesrvAddr string `json:"namesrv_addr" binding:"required,max=500"`
	AccessKey   string `json:"access_key" binding:"max=200"`
	SecretKey   string `json:"secret_key" binding:"max=200"`
	Enabled     *int   `json:"enabled"`
}

type TestConnectionRequest struct {
	Type        string `json:"type" binding:"required,oneof=rocketmq kafka rabbitmq"`
	NamesrvAddr string `json:"namesrv_addr" binding:"required,max=500"`
	AccessKey   string `json:"access_key"`
	SecretKey   string `json:"secret_key"`
}

// Add 新增数据源
func (s *MQSourceService) Add(req AddMQSourceRequest) (*models.MQSource, error) {
	id, err := models.AddMQSource(
		req.Name,
		req.Type,
		FormatNamesrvAddr(req.NamesrvAddr),
		req.AccessKey,
		req.SecretKey,
		"", // createdBy
	)
	if err != nil {
		return nil, err
	}

	source, err := models.GetMQSourceByID(id)
	if err != nil {
		return nil, err
	}

	return &source, nil
}

// Edit 编辑数据源
func (s *MQSourceService) Edit(id string, req EditMQSourceRequest) error {
	data := map[string]interface{}{
		"name":         req.Name,
		"type":         req.Type,
		"namesrv_addr": FormatNamesrvAddr(req.NamesrvAddr),
		"access_key":   req.AccessKey,
		"secret_key":   req.SecretKey,
	}

	if req.Enabled != nil {
		data["enabled"] = *req.Enabled
	}

	return models.UpdateMQSource(id, data)
}

// Delete 删除数据源
func (s *MQSourceService) Delete(id string) error {
	// 检查是否有绑定的订阅
	count, err := models.GetSubscriptionsTotal("", id, "", "")
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("该数据源还有 %d 个订阅，请先删除订阅", count)
	}

	return models.DeleteMQSource(id)
}

// GetByID 根据 ID 获取数据源
func (s *MQSourceService) GetByID(id string) (*models.MQSource, error) {
	source, err := models.GetMQSourceByID(id)
	if err != nil {
		return nil, err
	}
	return &source, nil
}

// GetAll 获取数据源列表
func (s *MQSourceService) GetAll(name, status, mqType string, page, pageSize int) ([]models.MQSource, int64, error) {
	sources, err := models.GetMQSources(page, pageSize, name, mqType, status)
	if err != nil {
		return nil, 0, err
	}

	total, err := models.GetMQSourcesTotal(name, mqType, status)
	if err != nil {
		return nil, 0, err
	}

	// 补充 binding_count 字段
	for i := range sources {
		count, _ := models.GetSubscriptionsTotal("", sources[i].ID, "", "")
		sources[i].BindingCount = int(count)
	}

	return sources, total, nil
}

// Count 获取总数
func (s *MQSourceService) Count(name, status, mqType string) (int64, error) {
	return models.GetMQSourcesTotal(name, mqType, status)
}

// TestConnection 测试连接
// TODO: 实现真实的 RocketMQ 连接测试
// 需要添加 RocketMQ 依赖后实现
func (s *MQSourceService) TestConnection() error {
	if s.Name == "" {
		return fmt.Errorf("数据源名称不能为空")
	}
	// 当前版本仅校验基本参数
	// 后续实现：创建临时 PullConsumer 测试连接
	return nil
}

// TestConnectionByID 根据 ID 测试连接
func (s *MQSourceService) TestConnectionByID(id string) error {
	source, err := models.GetMQSourceByID(id)
	if err != nil {
		return err
	}

	// 校验基本参数
	if source.NamesrvAddr == "" {
		return fmt.Errorf("NameServer 地址不能为空")
	}

	return s.testTCPConnectivity(id, source.NamesrvAddr)
}

// TestConnectionDirect 直接测试连接配置（不需要保存的数据源）
func (s *MQSourceService) TestConnectionDirect(_ string, namesrvAddr, _ string, _ string) error {
	// 校验基本参数
	if namesrvAddr == "" {
		return fmt.Errorf("NameServer 地址不能为空")
	}

	return s.testTCPConnectivity("", namesrvAddr)
}

// FormatNamesrvAddr 格式化 NameServer 地址
func FormatNamesrvAddr(addr string) string {
	addr = strings.TrimSpace(addr)
	// 移除前后空格和换行
	addr = strings.ReplaceAll(addr, "\n", "")
	addr = strings.ReplaceAll(addr, "\r", "")
	// 多个地址用分号分隔
	addr = strings.ReplaceAll(addr, ",", ";")
	addr = strings.ReplaceAll(addr, "，", ";")
	return addr
}

func ParseNamesrvAddrs(addr string) []string {
	formatted := FormatNamesrvAddr(addr)
	if formatted == "" {
		return []string{}
	}
	parts := strings.Split(formatted, ";")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func (s *MQSourceService) testTCPConnectivity(id, namesrvAddr string) error {
	addrs := ParseNamesrvAddrs(namesrvAddr)
	if len(addrs) == 0 {
		errMsg := "NameServer 地址不能为空"
		if id != "" {
			models.UpdateMQSourceTestStatus(id, "failed", errMsg)
		}
		return fmt.Errorf("%s", errMsg)
	}

	timeout := 3 * time.Second
	var errs []string
	for _, addr := range addrs {
		target, err := normalizeAddrForDial(addr)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", addr, err))
			continue
		}
		conn, err := net.DialTimeout("tcp", target, timeout)
		if err == nil {
			_ = conn.Close()
			if id != "" {
				models.UpdateMQSourceTestStatus(id, "success", "")
			}
			return nil
		}
		errs = append(errs, fmt.Sprintf("%s: %v", target, err))
	}

	errMsg := fmt.Sprintf("连接失败（TCP探测）: %s", strings.Join(errs, " | "))
	if id != "" {
		models.UpdateMQSourceTestStatus(id, "failed", errMsg)
	}
	return fmt.Errorf("%s", errMsg)
}

func normalizeAddrForDial(addr string) (string, error) {
	a := strings.TrimSpace(addr)
	if a == "" {
		return "", fmt.Errorf("地址为空")
	}
	// 支持用户输入 http://host:port
	if strings.Contains(a, "://") {
		u, err := url.Parse(a)
		if err != nil {
			return "", fmt.Errorf("地址格式错误")
		}
		if u.Host == "" {
			return "", fmt.Errorf("缺少 host:port")
		}
		a = u.Host
	}
	if _, _, err := net.SplitHostPort(a); err != nil {
		return "", fmt.Errorf("地址必须包含端口，例如 127.0.0.1:9876")
	}
	return a, nil
}
