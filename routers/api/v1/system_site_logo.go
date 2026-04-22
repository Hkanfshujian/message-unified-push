package v1

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"strings"
	"time"

	"ops-message-unified-push/pkg/app"
	"ops-message-unified-push/pkg/constant"
	"ops-message-unified-push/service/settings_service"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const siteLogoMaxSize = 2 * 1024 * 1024 // 2MB

// UploadSiteLogo 上传站点 logo
func UploadSiteLogo(c *gin.Context) {
	appG := app.Gin{C: c}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		appG.CResponse(http.StatusBadRequest, "请选择要上传的文件", nil)
		return
	}
	defer file.Close()

	if header.Size > siteLogoMaxSize {
		appG.CResponse(http.StatusBadRequest, "文件不能超过 2MB", nil)
		return
	}

	raw, err := io.ReadAll(io.LimitReader(file, siteLogoMaxSize+1))
	if err != nil {
		appG.CResponse(http.StatusBadRequest, "读取文件失败", nil)
		return
	}
	if int64(len(raw)) > siteLogoMaxSize {
		appG.CResponse(http.StatusBadRequest, "文件不能超过 2MB", nil)
		return
	}

	// 验证是否为有效图片
	_, format, err := image.DecodeConfig(bytes.NewReader(raw))
	if err != nil {
		appG.CResponse(http.StatusBadRequest, "文件不是有效的图片格式", nil)
		return
	}
	if format != "png" && format != "jpeg" && format != "gif" {
		appG.CResponse(http.StatusBadRequest, "仅支持 PNG、JPG、GIF 格式", nil)
		return
	}

	// 获取存储配置
	profiles, _, err := loadStorageProfiles()
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, fmt.Sprintf("获取存储配置失败：%s", err.Error()), nil)
		return
	}

	// 查找 S3 存储
	var profile *StorageProfile
	for i := range profiles {
		if profiles[i].Provider == "s3" && profiles[i].S3Endpoint != "" && profiles[i].Enabled {
			profile = &profiles[i]
			break
		}
	}
	if profile == nil {
		appG.CResponse(http.StatusBadRequest, "请先配置 S3 存储", nil)
		return
	}

	// 生成文件名
	objectKey := fmt.Sprintf("site-logo-%d.%s", time.Now().UnixMilli(), format)

	// 上传到 S3
	endpoint, secure := normalizeS3Endpoint(profile.S3Endpoint, profile.S3UseSSL)
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(profile.S3AccessKey, profile.S3SecretKey, ""),
		Secure: secure,
	})
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, fmt.Sprintf("连接存储服务失败：%s", err.Error()), nil)
		return
	}

	contentType := "image/" + format
	_, err = client.PutObject(c.Request.Context(), profile.S3Bucket, objectKey, bytes.NewReader(raw), int64(len(raw)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, fmt.Sprintf("上传文件失败：%s", err.Error()), nil)
		return
	}

	// 构造访问 URL
	var logoURL string
	if profile.S3PublicBaseURL != "" {
		logoURL = strings.TrimSuffix(profile.S3PublicBaseURL, "/") + "/" + objectKey
	} else {
		protocol := "http"
		if secure {
			protocol = "https"
		}
		logoURL = fmt.Sprintf("%s://%s/%s/%s", protocol, endpoint, profile.S3Bucket, objectKey)
	}

	// 保存到配置
	currentUser := c.GetString("username")
	settingService := settings_service.UserSettings{}
	err = settingService.EditSettings(constant.SiteSettingSectionName, "logo", logoURL, currentUser)
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, fmt.Sprintf("保存配置失败：%s", err.Error()), nil)
		return
	}

	// 清除缓存
	settings_service.ClearSiteConfigCache()

	appG.CResponse(http.StatusOK, "上传成功", map[string]string{
		"logo_url": logoURL,
	})
}

// ClearSiteLogo 恢复默认站点 logo
func ClearSiteLogo(c *gin.Context) {
	appG := app.Gin{C: c}

	// 恢复默认 logo 配置
	currentUser := c.GetString("username")
	settingService := settings_service.UserSettings{}
	defaultLogo := strings.TrimSpace(constant.SiteSiteDefaultValueMap["logo"])
	if defaultLogo == "" {
		appG.CResponse(http.StatusInternalServerError, "default logo is empty", nil)
		return
	}
	err := settingService.EditSettings(constant.SiteSettingSectionName, "logo", defaultLogo, currentUser)
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, fmt.Sprintf("恢复默认图标失败：%s", err.Error()), nil)
		return
	}

	// 清除缓存
	settings_service.ClearSiteConfigCache()

	appG.CResponse(http.StatusOK, "恢复默认图标成功", map[string]string{
		"logo": defaultLogo,
	})
}
