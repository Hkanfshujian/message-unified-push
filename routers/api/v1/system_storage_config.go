package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"ops-message-unified-push/pkg/app"
	"ops-message-unified-push/pkg/constant"
	"ops-message-unified-push/pkg/e"
	"ops-message-unified-push/service/settings_service"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var storageIDPattern = regexp.MustCompile(`^\d{8}$`)

// normalizeS3Endpoint 处理 S3 endpoint URL，返回纯主机名/IP和是否使用 SSL
func normalizeS3Endpoint(endpoint string, useSSL bool) (string, bool) {
	endpoint = strings.TrimSpace(endpoint)
	secure := useSSL
	
	if strings.HasPrefix(endpoint, "https://") {
		endpoint = strings.TrimPrefix(endpoint, "https://")
		secure = true
	} else if strings.HasPrefix(endpoint, "http://") {
		endpoint = strings.TrimPrefix(endpoint, "http://")
		secure = false
	}
	
	endpoint = strings.TrimSuffix(endpoint, "/")
	return endpoint, secure
}

type StorageProfile struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Provider          string `json:"provider"`
	Enabled           bool   `json:"enabled"`
	UploadFilePrefix  string `json:"upload_file_prefix"`
	LocalSubPath      string `json:"local_sub_path"`
	S3Endpoint        string `json:"s3_endpoint"`
	S3Region          string `json:"s3_region"`
	S3Bucket          string `json:"s3_bucket"`
	S3AccessKey       string `json:"s3_access_key"`
	S3SecretKey       string `json:"s3_secret_key"`
	S3UseSSL          bool   `json:"s3_use_ssl"`
	S3PublicBaseURL   string `json:"s3_public_base_url"`
	S3ProxyPublicRead bool   `json:"s3_proxy_public_read"`
	S3ObjectKeyPrefix string `json:"s3_object_key_prefix"`
}

type UpdateSystemStorageConfigReq struct {
	DefaultStorageID string           `json:"default_storage_id"`
	Profiles         []StorageProfile `json:"profiles"`
}

type TestSystemStorageConfigReq struct {
	Profile StorageProfile `json:"profile"`
}

type CreateSystemStorageLocalDirectoryReq struct {
	Path string `json:"path"`
	Name string `json:"name"`
}

type LocalDirectoryEntry struct {
	Name         string `json:"name"`
	RelativePath string `json:"relative_path"`
}

type S3ObjectDirectoryEntry struct {
	Name         string `json:"name"`
	RelativePath string `json:"relative_path"`
}

type S3ObjectFileEntry struct {
	Name         string `json:"name"`
	RelativePath string `json:"relative_path"`
	ObjectKey    string `json:"object_key"`
	Size         int64  `json:"size"`
	LastModified string `json:"last_modified"`
	PublicURL    string `json:"public_url"`
}

type LocalFileEntry struct {
	Name         string `json:"name"`
	RelativePath string `json:"relative_path"`
	Size         int64  `json:"size"`
	LastModified string `json:"last_modified"`
	PublicURL    string `json:"public_url"`
}

type DeleteSystemStorageFileReq struct {
	ProfileID    string `json:"profile_id"`
	ObjectKey    string `json:"object_key"`
	RelativePath string `json:"relative_path"`
}

const storageUploadMaxSize = 10 * 1024 * 1024
const defaultUploadFilePrefix = "upload"

func normalizeUploadFilePrefix(prefix string) string {
	raw := strings.TrimSpace(prefix)
	if raw == "" {
		return defaultUploadFilePrefix
	}
	builder := strings.Builder{}
	for _, ch := range raw {
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '-' || ch == '_' {
			builder.WriteRune(ch)
		}
	}
	normalized := strings.Trim(builder.String(), "-_")
	if normalized == "" {
		return defaultUploadFilePrefix
	}
	return normalized
}

func normalizeStorageProfile(input StorageProfile) StorageProfile {
	profile := input
	profile.ID = strings.TrimSpace(profile.ID)
	profile.Name = strings.TrimSpace(profile.Name)
	profile.Provider = strings.ToLower(strings.TrimSpace(profile.Provider))
	profile.UploadFilePrefix = normalizeUploadFilePrefix(profile.UploadFilePrefix)
	profile.LocalSubPath = normalizeLocalSubPath(profile.LocalSubPath)
	profile.S3Endpoint = strings.TrimSpace(profile.S3Endpoint)
	profile.S3Region = strings.TrimSpace(profile.S3Region)
	profile.S3Bucket = strings.TrimSpace(profile.S3Bucket)
	profile.S3AccessKey = strings.TrimSpace(profile.S3AccessKey)
	profile.S3SecretKey = strings.TrimSpace(profile.S3SecretKey)
	profile.S3PublicBaseURL = strings.TrimRight(strings.TrimSpace(profile.S3PublicBaseURL), "/")
	profile.S3ObjectKeyPrefix = strings.Trim(strings.TrimSpace(profile.S3ObjectKeyPrefix), "/")
	if profile.Provider == "" {
		profile.Provider = "local"
	}
	if profile.ID == "" {
		profile.ID = generateEightDigitStorageID(time.Now().UnixNano())
	}
	if profile.Name == "" {
		if profile.Provider == "s3" {
			profile.Name = "S3 存储"
		} else {
			profile.Name = "本地存储"
		}
	}
	if profile.Provider == "s3" {
	} else {
		if profile.LocalSubPath == "" {
			profile.LocalSubPath = "uploads"
		}
		profile.S3UseSSL = true
		profile.S3ProxyPublicRead = true
	}
	return profile
}

func normalizeLocalSubPath(input string) string {
	normalized := strings.ReplaceAll(strings.TrimSpace(input), "\\", "/")
	normalized = strings.Trim(normalized, "/")
	if normalized == "" {
		return ""
	}
	parts := strings.Split(normalized, "/")
	validParts := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" || part == "." || part == ".." {
			continue
		}
		validParts = append(validParts, part)
	}
	return strings.Join(validParts, "/")
}

func buildLocalStorageAbsPath(localSubPath string) string {
	subPath := normalizeLocalSubPath(localSubPath)
	if subPath == "" {
		subPath = "uploads"
	}
	return filepath.Join(".", "data", filepath.FromSlash(subPath))
}

func buildDataRootAbsPath() string {
	return filepath.Join(".", "data")
}

func buildDataChildAbsPath(relativePath string) string {
	child := normalizeLocalSubPath(relativePath)
	if child == "" {
		return buildDataRootAbsPath()
	}
	return filepath.Join(buildDataRootAbsPath(), filepath.FromSlash(child))
}

func ensureLocalStorageDir(localSubPath string) error {
	absPath := buildLocalStorageAbsPath(localSubPath)
	return os.MkdirAll(absPath, 0755)
}

func parseBoolFormValue(value string) bool {
	v := strings.ToLower(strings.TrimSpace(value))
	return v == "1" || v == "true" || v == "yes" || v == "on"
}

func buildS3PublicURL(baseURL, objectKey string) string {
	base := strings.TrimSpace(baseURL)
	key := strings.TrimLeft(strings.TrimSpace(objectKey), "/")
	if base == "" || key == "" {
		return ""
	}
	return strings.TrimRight(base, "/") + "/" + key
}

func isEightDigitStorageID(storageID string) bool {
	return storageIDPattern.MatchString(strings.TrimSpace(storageID))
}

func generateEightDigitStorageID(seed int64) string {
	base := int(seed % 90000000)
	if base < 0 {
		base = -base
	}
	return fmt.Sprintf("%08d", base+10000000)
}

func normalizeStorageIDs(profiles []StorageProfile, defaultStorageID string) ([]StorageProfile, string) {
	used := map[string]bool{}
	for i := range profiles {
		id := strings.TrimSpace(profiles[i].ID)
		if !isEightDigitStorageID(id) || used[id] {
			seed := time.Now().UnixNano() + int64(i*37)
			id = generateEightDigitStorageID(seed)
			for used[id] {
				seed += 97
				id = generateEightDigitStorageID(seed)
			}
		}
		profiles[i].ID = id
		used[id] = true
	}
	defaultID := strings.TrimSpace(defaultStorageID)
	if !isEightDigitStorageID(defaultID) || !used[defaultID] {
		if len(profiles) > 0 {
			defaultID = profiles[0].ID
		} else {
			defaultID = ""
		}
	}
	return profiles, defaultID
}

func validateStorageProfile(profile StorageProfile) string {
	if profile.Provider != "local" && profile.Provider != "s3" {
		return "存储类型仅支持 local 或 s3"
	}
	if !isEightDigitStorageID(profile.ID) {
		return "存储ID必须为8位数字"
	}
	if profile.Name == "" {
		return "存储名称不能为空"
	}
	if profile.Provider == "local" {
		if strings.TrimSpace(profile.LocalSubPath) == "" {
			return "本地存储路径不能为空"
		}
	}
	if profile.Provider == "s3" {
		if profile.S3Endpoint == "" || profile.S3Bucket == "" || profile.S3AccessKey == "" || profile.S3SecretKey == "" || profile.S3ObjectKeyPrefix == "" {
			return "S3 storage requires Endpoint, Bucket, Object Prefix, Access Key, Secret Key"
		}
	}
	return ""
}

func loadStorageProfiles() ([]StorageProfile, string, error) {
	settingService := settings_service.UserSettings{}
	data, err := settingService.GetUserSetting(constant.StorageConfigSectionName)
	if err != nil {
		return nil, "", err
	}
	defaultStorageID := strings.TrimSpace(data["default_storage_id"])
	profilesRaw := strings.TrimSpace(data["storage_profiles_json"])
	profiles := make([]StorageProfile, 0)
	if profilesRaw != "" {
		_ = json.Unmarshal([]byte(profilesRaw), &profiles)
	}
	if len(profiles) == 0 {
		legacy := normalizeStorageProfile(StorageProfile{
			ID:                "10000001",
			Name:              "默认存储",
			Provider:          data["storage_provider"],
			Enabled:           true,
			UploadFilePrefix:  data["upload_file_prefix"],
			S3Endpoint:        data["s3_endpoint"],
			S3Region:          data["s3_region"],
			S3Bucket:          data["s3_bucket"],
			S3AccessKey:       data["s3_access_key"],
			S3SecretKey:       data["s3_secret_key"],
			LocalSubPath:      data["local_sub_path"],
			S3UseSSL:          strings.EqualFold(data["s3_use_ssl"], "true") || data["s3_use_ssl"] == "1",
			S3PublicBaseURL:   data["s3_public_base_url"],
			S3ProxyPublicRead: strings.EqualFold(data["s3_proxy_public_read"], "true") || data["s3_proxy_public_read"] == "1",
			S3ObjectKeyPrefix: data["s3_object_key_prefix"],
		})
		profiles = append(profiles, legacy)
		if defaultStorageID == "" {
			defaultStorageID = legacy.ID
		}
	}
	for i := range profiles {
		profiles[i] = normalizeStorageProfile(profiles[i])
	}
	profiles, defaultStorageID = normalizeStorageIDs(profiles, defaultStorageID)
	if defaultStorageID == "" && len(profiles) > 0 {
		defaultStorageID = profiles[0].ID
	}
	return profiles, defaultStorageID, nil
}

func saveStorageProfiles(currentUser string, profiles []StorageProfile, defaultStorageID string) error {
	normalizedProfiles := make([]StorageProfile, 0, len(profiles))
	for _, profile := range profiles {
		normalizedProfiles = append(normalizedProfiles, normalizeStorageProfile(profile))
	}
	if defaultStorageID == "" && len(normalizedProfiles) > 0 {
		defaultStorageID = normalizedProfiles[0].ID
	}
	raw, err := json.Marshal(normalizedProfiles)
	if err != nil {
		return err
	}
	settingService := settings_service.UserSettings{}
	if err = settingService.EditSettings(constant.StorageConfigSectionName, "storage_profiles_json", string(raw), currentUser); err != nil {
		return err
	}
	if err = settingService.EditSettings(constant.StorageConfigSectionName, "default_storage_id", defaultStorageID, currentUser); err != nil {
		return err
	}
	return nil
}

func findStorageProfileByID(profiles []StorageProfile, profileID string) (StorageProfile, bool) {
	target := strings.TrimSpace(profileID)
	for _, profile := range profiles {
		if profile.ID == target {
			return profile, true
		}
	}
	return StorageProfile{}, false
}

func ResolveDefaultStorageProfile() (StorageProfile, error) {
	profiles, defaultStorageID, err := loadStorageProfiles()
	if err != nil {
		return StorageProfile{}, err
	}
	profile, ok := findStorageProfileByID(profiles, defaultStorageID)
	if !ok {
		if len(profiles) == 0 {
			return StorageProfile{}, fmt.Errorf("未配置可用存储")
		}
		return profiles[0], nil
	}
	return profile, nil
}

func ResolveStorageProfileByID(profileID string) (StorageProfile, error) {
	profiles, _, err := loadStorageProfiles()
	if err != nil {
		return StorageProfile{}, err
	}
	profile, ok := findStorageProfileByID(profiles, profileID)
	if !ok {
		return StorageProfile{}, fmt.Errorf("存储配置不存在")
	}
	return profile, nil
}

func GetSystemStorageConfig(c *gin.Context) {
	appG := app.Gin{C: c}
	profiles, defaultStorageID, err := loadStorageProfiles()
	if err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("获取存储配置失败：%s", err.Error()), nil)
		return
	}
	appG.CResponse(http.StatusOK, "获取存储配置成功", map[string]interface{}{
		"default_storage_id": defaultStorageID,
		"profiles":           profiles,
	})
}

func GetSystemStorageLocalDirectories(c *gin.Context) {
	appG := app.Gin{C: c}
	currentPath := normalizeLocalSubPath(c.Query("path"))
	rootAbs := buildDataRootAbsPath()
	targetAbs := buildDataChildAbsPath(currentPath)
	if err := os.MkdirAll(rootAbs, 0755); err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("初始化本地根目录失败：%s", err.Error()), nil)
		return
	}
	rel, err := filepath.Rel(rootAbs, targetAbs)
	if err != nil || strings.HasPrefix(rel, "..") {
		appG.CResponse(http.StatusBadRequest, "目录路径非法", nil)
		return
	}
	entries, err := os.ReadDir(targetAbs)
	if err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("读取目录失败：%s", err.Error()), nil)
		return
	}
	dirs := make([]LocalDirectoryEntry, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dirName := strings.TrimSpace(entry.Name())
		if dirName == "" || dirName == "." || dirName == ".." {
			continue
		}
		childRelativePath := normalizeLocalSubPath(filepath.ToSlash(filepath.Join(currentPath, dirName)))
		dirs = append(dirs, LocalDirectoryEntry{
			Name:         dirName,
			RelativePath: childRelativePath,
		})
	}
	sort.Slice(dirs, func(i, j int) bool {
		return strings.ToLower(dirs[i].Name) < strings.ToLower(dirs[j].Name)
	})
	parentPath := ""
	if currentPath != "" {
		idx := strings.LastIndex(currentPath, "/")
		if idx >= 0 {
			parentPath = currentPath[:idx]
		}
	}
	appG.CResponse(http.StatusOK, "读取本地目录成功", map[string]interface{}{
		"root":         "./data",
		"current_path": currentPath,
		"parent_path":  parentPath,
		"directories":  dirs,
	})
}

func GetSystemStorageLocalFiles(c *gin.Context) {
	appG := app.Gin{C: c}
	profileID := strings.TrimSpace(c.Query("profile_id"))
	if profileID == "" {
		appG.CResponse(http.StatusBadRequest, "缺少存储ID", nil)
		return
	}
	path := normalizeLocalSubPath(c.Query("path"))
	profile, err := ResolveStorageProfileByID(profileID)
	if err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("读取存储配置失败：%s", err.Error()), nil)
		return
	}
	profile = normalizeStorageProfile(profile)
	if profile.Provider != "local" {
		appG.CResponse(http.StatusBadRequest, "仅支持本地存储浏览", nil)
		return
	}
	rootPath := normalizeLocalSubPath(profile.LocalSubPath)
	if rootPath == "" {
		rootPath = "uploads"
	}
	currentPath := normalizeLocalSubPath(path)
	absoluteCurrentPath := rootPath
	if currentPath != "" {
		absoluteCurrentPath = normalizeLocalSubPath(rootPath + "/" + currentPath)
	}
	rootAbs := buildDataRootAbsPath()
	targetAbs := buildDataChildAbsPath(absoluteCurrentPath)
	if err = os.MkdirAll(buildDataChildAbsPath(rootPath), 0755); err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("初始化本地目录失败：%s", err.Error()), nil)
		return
	}
	rel, err := filepath.Rel(rootAbs, targetAbs)
	if err != nil || strings.HasPrefix(rel, "..") {
		appG.CResponse(http.StatusBadRequest, "目录路径非法", nil)
		return
	}
	entries, err := os.ReadDir(targetAbs)
	if err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("读取目录失败：%s", err.Error()), nil)
		return
	}
	directories := make([]LocalDirectoryEntry, 0)
	files := make([]LocalFileEntry, 0)
	for _, entry := range entries {
		name := strings.TrimSpace(entry.Name())
		if name == "" || name == "." || name == ".." {
			continue
		}
		nextFullRelative := normalizeLocalSubPath(filepath.ToSlash(filepath.Join(absoluteCurrentPath, name)))
		nextRelativePath := strings.TrimPrefix(nextFullRelative, rootPath)
		nextRelativePath = strings.TrimPrefix(nextRelativePath, "/")
		if entry.IsDir() {
			directories = append(directories, LocalDirectoryEntry{
				Name:         name,
				RelativePath: nextRelativePath,
			})
			continue
		}
		info, infoErr := entry.Info()
		if infoErr != nil {
			continue
		}
		files = append(files, LocalFileEntry{
			Name:         name,
			RelativePath: nextRelativePath,
			Size:         info.Size(),
			LastModified: info.ModTime().Format("2006-01-02 15:04:05"),
			PublicURL:    buildLocalPublicURL(nextFullRelative),
		})
	}
	sort.Slice(directories, func(i, j int) bool {
		return strings.ToLower(directories[i].Name) < strings.ToLower(directories[j].Name)
	})
	sort.Slice(files, func(i, j int) bool {
		return strings.ToLower(files[i].Name) < strings.ToLower(files[j].Name)
	})
	parentPath := ""
	if currentPath != "" {
		if idx := strings.LastIndex(currentPath, "/"); idx >= 0 {
			parentPath = currentPath[:idx]
		}
	}
	appG.CResponse(http.StatusOK, "读取本地文件成功", map[string]interface{}{
		"profile_id":   profile.ID,
		"profile_name": profile.Name,
		"root_path":    rootPath,
		"current_path": currentPath,
		"parent_path":  parentPath,
		"directories":  directories,
		"files":        files,
	})
}

func GetSystemStorageS3Objects(c *gin.Context) {
	appG := app.Gin{C: c}
	profileID := strings.TrimSpace(c.Query("profile_id"))
	if profileID == "" {
		appG.CResponse(http.StatusBadRequest, "缺少存储ID", nil)
		return
	}
	path := normalizeLocalSubPath(c.Query("path"))
	profile, err := ResolveStorageProfileByID(profileID)
	if err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("读取存储配置失败：%s", err.Error()), nil)
		return
	}
	profile = normalizeStorageProfile(profile)
	if profile.Provider != "s3" {
		appG.CResponse(http.StatusBadRequest, "仅支持 S3 存储浏览", nil)
		return
	}
	basePrefix := strings.Trim(profile.S3ObjectKeyPrefix, "/")
	if basePrefix == "" {
		appG.CResponse(http.StatusBadRequest, "S3 对象前缀不能为空", nil)
		return
	}
	fullPrefix := basePrefix
	if path != "" {
		fullPrefix = basePrefix + "/" + path
	}
	if fullPrefix != "" {
		fullPrefix += "/"
	}
	endpoint, secure := normalizeS3Endpoint(profile.S3Endpoint, profile.S3UseSSL)
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(profile.S3AccessKey, profile.S3SecretKey, ""),
		Secure: secure,
		Region: profile.S3Region,
	})
	if err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("初始化 S3 客户端失败：%s", err.Error()), nil)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 18*time.Second)
	defer cancel()
	result := client.ListObjects(ctx, profile.S3Bucket, minio.ListObjectsOptions{
		Prefix:    fullPrefix,
		Recursive: false,
		MaxKeys:   200,
	})
	directories := make([]S3ObjectDirectoryEntry, 0)
	files := make([]S3ObjectFileEntry, 0)
	for object := range result {
		if object.Err != nil {
			appG.CResponse(http.StatusBadRequest, fmt.Sprintf("读取 S3 对象失败：%s", object.Err.Error()), nil)
			return
		}
		rawKey := strings.TrimSpace(object.Key)
		key := strings.Trim(rawKey, "/")
		if key == "" || key == basePrefix {
			continue
		}
		relative := strings.TrimPrefix(key, basePrefix)
		relative = strings.TrimPrefix(relative, "/")
		if relative == "" {
			continue
		}
		name := relative
		if idx := strings.LastIndex(relative, "/"); idx >= 0 {
			name = relative[idx+1:]
		}
		if strings.HasSuffix(rawKey, "/") {
			directories = append(directories, S3ObjectDirectoryEntry{
				Name:         strings.TrimSuffix(name, "/"),
				RelativePath: strings.TrimSuffix(relative, "/"),
			})
			continue
		}
		files = append(files, S3ObjectFileEntry{
			Name:         name,
			RelativePath: relative,
			ObjectKey:    key,
			Size:         object.Size,
			LastModified: object.LastModified.Format("2006-01-02 15:04:05"),
			PublicURL:    buildS3PublicURL(profile.S3PublicBaseURL, key),
		})
	}
	sort.Slice(directories, func(i, j int) bool {
		return strings.ToLower(directories[i].Name) < strings.ToLower(directories[j].Name)
	})
	sort.Slice(files, func(i, j int) bool {
		return strings.ToLower(files[i].Name) < strings.ToLower(files[j].Name)
	})
	parentPath := ""
	if path != "" {
		if idx := strings.LastIndex(path, "/"); idx >= 0 {
			parentPath = path[:idx]
		}
	}
	appG.CResponse(http.StatusOK, "读取 S3 对象成功", map[string]interface{}{
		"profile_id":     profile.ID,
		"profile_name":   profile.Name,
		"bucket":         profile.S3Bucket,
		"prefix":         basePrefix,
		"current_path":   path,
		"parent_path":    parentPath,
		"directories":    directories,
		"files":          files,
		"has_public_url": strings.TrimSpace(profile.S3PublicBaseURL) != "",
	})
}

func buildLocalPublicURL(relative string) string {
	normalized := strings.Trim(strings.TrimSpace(relative), "/")
	if normalized == "" {
		return "/storage"
	}
	if strings.HasPrefix(normalized, "uploads/") {
		return "/uploads/" + strings.TrimPrefix(normalized, "uploads/")
	}
	return "/storage/" + normalized
}

func DeleteSystemStorageFile(c *gin.Context) {
	var req DeleteSystemStorageFileReq
	appG := app.Gin{C: c}
	errCode, errMsg := app.BindJsonAndPlayValid(c, &req)
	if errCode != e.SUCCESS {
		appG.CResponse(errCode, errMsg, nil)
		return
	}
	profileID := strings.TrimSpace(req.ProfileID)
	if profileID == "" {
		appG.CResponse(http.StatusBadRequest, "缺少存储ID", nil)
		return
	}
	profile, err := ResolveStorageProfileByID(profileID)
	if err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("读取存储配置失败：%s", err.Error()), nil)
		return
	}
	profile = normalizeStorageProfile(profile)
	if profile.Provider == "s3" {
		objectKey := strings.Trim(strings.TrimSpace(req.ObjectKey), "/")
		if objectKey == "" {
			appG.CResponse(http.StatusBadRequest, "缺少对象键", nil)
			return
		}
		basePrefix := strings.Trim(profile.S3ObjectKeyPrefix, "/")
		if basePrefix != "" && objectKey != basePrefix && !strings.HasPrefix(objectKey, basePrefix+"/") {
			appG.CResponse(http.StatusBadRequest, "对象键超出存储前缀范围", nil)
			return
		}
		endpoint, secure := normalizeS3Endpoint(profile.S3Endpoint, profile.S3UseSSL)
		client, err := minio.New(endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(profile.S3AccessKey, profile.S3SecretKey, ""),
			Secure: secure,
			Region: profile.S3Region,
		})
		if err != nil {
			appG.CResponse(http.StatusBadRequest, fmt.Sprintf("初始化 S3 客户端失败：%s", err.Error()), nil)
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), 18*time.Second)
		defer cancel()
		if err = client.RemoveObject(ctx, profile.S3Bucket, objectKey, minio.RemoveObjectOptions{}); err != nil {
			appG.CResponse(http.StatusBadRequest, fmt.Sprintf("删除 S3 对象失败：%s", err.Error()), nil)
			return
		}
		appG.CResponse(http.StatusOK, "删除成功", map[string]string{
			"provider":   "s3",
			"object_key": objectKey,
		})
		return
	}
	relativePath := normalizeLocalSubPath(req.RelativePath)
	if relativePath == "" {
		appG.CResponse(http.StatusBadRequest, "缺少文件相对路径", nil)
		return
	}
	rootPath := normalizeLocalSubPath(profile.LocalSubPath)
	if rootPath == "" {
		rootPath = "uploads"
	}
	fullRelative := normalizeLocalSubPath(filepath.ToSlash(filepath.Join(rootPath, relativePath)))
	if fullRelative == "" || (fullRelative != rootPath && !strings.HasPrefix(fullRelative, rootPath+"/")) {
		appG.CResponse(http.StatusBadRequest, "文件路径非法", nil)
		return
	}
	rootAbs := buildDataRootAbsPath()
	targetAbs := buildDataChildAbsPath(fullRelative)
	rel, err := filepath.Rel(rootAbs, targetAbs)
	if err != nil || strings.HasPrefix(rel, "..") {
		appG.CResponse(http.StatusBadRequest, "文件路径非法", nil)
		return
	}
	info, err := os.Stat(targetAbs)
	if err != nil {
		if os.IsNotExist(err) {
			appG.CResponse(http.StatusOK, "文件不存在，视为删除成功", map[string]string{
				"provider":      "local",
				"relative_path": relativePath,
			})
			return
		}
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("读取文件失败：%s", err.Error()), nil)
		return
	}
	if info.IsDir() {
		appG.CResponse(http.StatusBadRequest, "当前路径是目录，无法直接删除", nil)
		return
	}
	if err = os.Remove(targetAbs); err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("删除本地文件失败：%s", err.Error()), nil)
		return
	}
	appG.CResponse(http.StatusOK, "删除成功", map[string]string{
		"provider":      "local",
		"relative_path": relativePath,
	})
}

func CreateSystemStorageLocalDirectory(c *gin.Context) {
	var req CreateSystemStorageLocalDirectoryReq
	appG := app.Gin{C: c}
	errCode, errMsg := app.BindJsonAndPlayValid(c, &req)
	if errCode != e.SUCCESS {
		appG.CResponse(errCode, errMsg, nil)
		return
	}
	parentPath := normalizeLocalSubPath(req.Path)
	newDirName := strings.TrimSpace(req.Name)
	if newDirName == "" {
		appG.CResponse(http.StatusBadRequest, "目录名称不能为空", nil)
		return
	}
	if strings.Contains(newDirName, "/") || strings.Contains(newDirName, "\\") || strings.Contains(newDirName, "..") {
		appG.CResponse(http.StatusBadRequest, "目录名称非法", nil)
		return
	}
	rootAbs := buildDataRootAbsPath()
	if err := os.MkdirAll(rootAbs, 0755); err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("初始化本地根目录失败：%s", err.Error()), nil)
		return
	}
	newRelativePath := normalizeLocalSubPath(filepath.ToSlash(filepath.Join(parentPath, newDirName)))
	targetAbs := buildDataChildAbsPath(newRelativePath)
	rel, err := filepath.Rel(rootAbs, targetAbs)
	if err != nil || strings.HasPrefix(rel, "..") {
		appG.CResponse(http.StatusBadRequest, "目录路径非法", nil)
		return
	}
	if err = os.MkdirAll(targetAbs, 0755); err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("创建目录失败：%s", err.Error()), nil)
		return
	}
	appG.CResponse(http.StatusOK, "创建目录成功", map[string]string{
		"relative_path": newRelativePath,
	})
}

func UploadSystemStorageFile(c *gin.Context) {
	appG := app.Gin{C: c}
	profileID := strings.TrimSpace(c.PostForm("profile_id"))
	deleteAfterUpload := parseBoolFormValue(c.PostForm("delete_after_upload"))
	if profileID == "" {
		appG.CResponse(http.StatusBadRequest, "缺少存储ID", nil)
		return
	}
	profile, err := ResolveStorageProfileByID(profileID)
	if err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("读取存储配置失败：%s", err.Error()), nil)
		return
	}
	profile = normalizeStorageProfile(profile)
	if profile.Provider != "local" && profile.Provider != "s3" {
		appG.CResponse(http.StatusBadRequest, "仅支持本地或S3存储上传", nil)
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		appG.CResponse(http.StatusBadRequest, "请选择要上传的文件", nil)
		return
	}
	defer file.Close()
	if header.Size > storageUploadMaxSize {
		appG.CResponse(http.StatusBadRequest, "文件不能超过 10MB", nil)
		return
	}

	raw, err := io.ReadAll(io.LimitReader(file, storageUploadMaxSize+1))
	if err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("读取文件失败：%s", err.Error()), nil)
		return
	}
	if int64(len(raw)) > storageUploadMaxSize {
		appG.CResponse(http.StatusBadRequest, "文件不能超过 10MB", nil)
		return
	}

	contentType := http.DetectContentType(raw)
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext == "" || len(ext) > 10 {
		ext = ".bin"
	}
	now := time.Now()
	uploadFileName := fmt.Sprintf("%s-%d%s", profile.UploadFilePrefix, now.UnixNano(), ext)
	if profile.Provider == "local" {
		targetDir := buildLocalStorageAbsPath(profile.LocalSubPath)
		if err = os.MkdirAll(targetDir, 0755); err != nil {
			appG.CResponse(http.StatusBadRequest, fmt.Sprintf("创建本地目录失败：%s", err.Error()), nil)
			return
		}
		month := now.Format("200601")
		uploadDir := filepath.Join(targetDir, "manual-upload", month)
		if err = os.MkdirAll(uploadDir, 0755); err != nil {
			appG.CResponse(http.StatusBadRequest, fmt.Sprintf("创建测试目录失败：%s", err.Error()), nil)
			return
		}
		uploadPath := filepath.Join(uploadDir, uploadFileName)
		if err = os.WriteFile(uploadPath, raw, 0644); err != nil {
			appG.CResponse(http.StatusBadRequest, fmt.Sprintf("写入测试文件失败：%s", err.Error()), nil)
			return
		}
		if deleteAfterUpload {
			_ = os.Remove(uploadPath)
		}
		relPath := normalizeLocalSubPath(filepath.ToSlash(filepath.Join(profile.LocalSubPath, "manual-upload", month, uploadFileName)))
		appG.CResponse(http.StatusOK, "本地文件上传成功", map[string]string{
			"provider":      "local",
			"status":        "ok",
			"relative_path": relPath,
			"deleted":       fmt.Sprintf("%t", deleteAfterUpload),
		})
		return
	}
	endpoint, secure := normalizeS3Endpoint(profile.S3Endpoint, profile.S3UseSSL)
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(profile.S3AccessKey, profile.S3SecretKey, ""),
		Secure: secure,
		Region: profile.S3Region,
	})
	if err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("初始化 S3 客户端失败：%s", err.Error()), nil)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 18*time.Second)
	defer cancel()
	objectKey := fmt.Sprintf("%s/%s", strings.Trim(profile.S3ObjectKeyPrefix, "/"), uploadFileName)
	if _, err = client.PutObject(ctx, profile.S3Bucket, objectKey, bytes.NewReader(raw), int64(len(raw)), minio.PutObjectOptions{
		ContentType: contentType,
	}); err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("上传 S3 测试文件失败：%s", err.Error()), nil)
		return
	}
	if deleteAfterUpload {
		_ = client.RemoveObject(ctx, profile.S3Bucket, objectKey, minio.RemoveObjectOptions{})
	}
	appG.CResponse(http.StatusOK, "S3 文件上传成功", map[string]string{
		"provider":   "s3",
		"status":     "ok",
		"bucket":     profile.S3Bucket,
		"object_key": objectKey,
		"public_url": buildS3PublicURL(profile.S3PublicBaseURL, objectKey),
		"deleted":    fmt.Sprintf("%t", deleteAfterUpload),
	})
}

func UpdateSystemStorageConfig(c *gin.Context) {
	var req UpdateSystemStorageConfigReq
	appG := app.Gin{C: c}
	errCode, errMsg := app.BindJsonAndPlayValid(c, &req)
	if errCode != e.SUCCESS {
		appG.CResponse(errCode, errMsg, nil)
		return
	}
	if len(req.Profiles) == 0 {
		appG.CResponse(http.StatusBadRequest, "至少保留一个存储配置", nil)
		return
	}
	idSet := map[string]bool{}
	for _, profile := range req.Profiles {
		normalized := normalizeStorageProfile(profile)
		if idSet[normalized.ID] {
			appG.CResponse(http.StatusBadRequest, "存储ID不能重复", nil)
			return
		}
		idSet[normalized.ID] = true
		if msg := validateStorageProfile(normalized); msg != "" {
			appG.CResponse(http.StatusBadRequest, msg, nil)
			return
		}
		if normalized.Provider == "local" {
			if err := ensureLocalStorageDir(normalized.LocalSubPath); err != nil {
				appG.CResponse(http.StatusBadRequest, fmt.Sprintf("创建本地存储目录失败：%s", err.Error()), nil)
				return
			}
		}
	}
	if strings.TrimSpace(req.DefaultStorageID) == "" {
		req.DefaultStorageID = req.Profiles[0].ID
	}
	if !isEightDigitStorageID(req.DefaultStorageID) {
		appG.CResponse(http.StatusBadRequest, "默认存储ID必须为8位数字", nil)
		return
	}
	if !idSet[strings.TrimSpace(req.DefaultStorageID)] {
		appG.CResponse(http.StatusBadRequest, "默认存储不存在", nil)
		return
	}
	currentUser := app.GetCurrentUserName(c)
	if err := saveStorageProfiles(currentUser, req.Profiles, req.DefaultStorageID); err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("保存存储配置失败：%s", err.Error()), nil)
		return
	}
	appG.CResponse(http.StatusOK, "保存存储配置成功", nil)
}

func TestSystemStorageConfig(c *gin.Context) {
	var req TestSystemStorageConfigReq
	appG := app.Gin{C: c}
	if c.Request.ContentLength > 0 {
		errCode, errMsg := app.BindJsonAndPlayValid(c, &req)
		if errCode != e.SUCCESS {
			appG.CResponse(errCode, errMsg, nil)
			return
		}
	}

	profile := normalizeStorageProfile(req.Profile)
	if profile.ID == "" {
		defaultProfile, err := ResolveDefaultStorageProfile()
		if err != nil {
			appG.CResponse(http.StatusBadRequest, fmt.Sprintf("读取默认存储失败：%s", err.Error()), nil)
			return
		}
		profile = defaultProfile
	}
	if msg := validateStorageProfile(profile); msg != "" {
		appG.CResponse(http.StatusBadRequest, msg, nil)
		return
	}
	if profile.Provider == "local" {
		appG.CResponse(http.StatusOK, "本地存储无需连通性测试", map[string]string{
			"provider": "local",
			"status":   "ok",
		})
		return
	}
	endpoint, secure := normalizeS3Endpoint(profile.S3Endpoint, profile.S3UseSSL)
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(profile.S3AccessKey, profile.S3SecretKey, ""),
		Secure: secure,
		Region: profile.S3Region,
	})
	if err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("初始化 S3 客户端失败：%s", err.Error()), nil)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 12*time.Second)
	defer cancel()
	exists, err := client.BucketExists(ctx, profile.S3Bucket)
	if err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("检查 Bucket 失败：%s", err.Error()), nil)
		return
	}
	if !exists {
		appG.CResponse(http.StatusBadRequest, "S3 Bucket 不存在或不可访问", nil)
		return
	}

	testKey := buildStorageTestObjectKey(profile.S3ObjectKeyPrefix)
	_, err = client.PutObject(ctx, profile.S3Bucket, testKey, bytes.NewReader([]byte("oidc-icon-storage-test")), int64(len("oidc-icon-storage-test")), minio.PutObjectOptions{
		ContentType: "text/plain",
	})
	if err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("写入测试对象失败：%s", err.Error()), nil)
		return
	}
	_ = client.RemoveObject(ctx, profile.S3Bucket, testKey, minio.RemoveObjectOptions{})

	appG.CResponse(http.StatusOK, "S3 连通性测试通过", map[string]string{
		"provider": "s3",
		"bucket":   profile.S3Bucket,
		"status":   "ok",
	})
}

func buildStorageTestObjectKey(prefix string) string {
	now := time.Now()
	trimPrefix := strings.Trim(prefix, "/")
	return fmt.Sprintf("%s/connectivity-test/%s/test-%d.txt", trimPrefix, now.Format("200601"), now.UnixNano())
}
