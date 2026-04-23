package setting

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

var optionValueMap = map[string]string{}

func getEnvValue(keys ...string) (string, bool) {
	for _, key := range keys {
		value := os.Getenv(key)
		if value != "" {
			optionValueMap[key] = value
			return value, true
		}
	}
	return "", false
}

func applyStringEnv(target *string, keys ...string) {
	if value, ok := getEnvValue(keys...); ok {
		*target = value
	}
}

func applyIntEnv(target *int, keys ...string) {
	if value, ok := getEnvValue(keys...); ok {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			log.Printf("[ops-message-unified-push] invalid int env value for %v: %s", keys, value)
			return
		}
		*target = parsed
	}
}

func applyDurationSecondsEnv(target *time.Duration, keys ...string) {
	if value, ok := getEnvValue(keys...); ok {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			log.Printf("[ops-message-unified-push] invalid duration env value for %v: %s", keys, value)
			return
		}
		*target = time.Duration(parsed)
	}
}

func initDefaultSettings() {
	AppSetting.JwtSecret = "ops-message-unified-push"
	AppSetting.RuntimeRootPath = "runtime/"
	AppSetting.LogLevel = "INFO"
	AppSetting.InitData = ""

	ServerSetting.RunMode = "release"
	ServerSetting.HttpPort = 8081
	ServerSetting.ReadTimeout = time.Duration(60)
	ServerSetting.WriteTimeout = time.Duration(60)
	ServerSetting.EmbedHtml = ""
	ServerSetting.UrlPrefix = ""

	DatabaseSetting.Type = "sqlite"
	DatabaseSetting.User = ""
	DatabaseSetting.Password = ""
	DatabaseSetting.Host = ""
	DatabaseSetting.Port = 3306
	DatabaseSetting.Name = ""
	DatabaseSetting.TablePrefix = "message_"
	DatabaseSetting.SqlDebug = "disable"
	DatabaseSetting.Ssl = "false"
}

var sensitiveKeys = []string{"PASSWORD", "SECRET", "KEY", "TOKEN", "CREDENTIAL"}

func printOptionValue() {
	for key, val := range optionValueMap {
		isSensitive := false
		upperKey := strings.ToUpper(key)
		for _, sk := range sensitiveKeys {
			if strings.Contains(upperKey, sk) {
				isSensitive = true
				break
			}
		}
		if isSensitive {
			log.Printf("[ops-message-unified-push] current option env: %s, value: *****", key)
		} else {
			log.Printf("[ops-message-unified-push] current option env: %s, value: %s", key, val)
		}
	}
}

func applyEnvOverrides() {
	applyStringEnv(&AppSetting.JwtSecret, "APP_JWT_SECRET", "JWT_SECRET")
	applyStringEnv(&AppSetting.RuntimeRootPath, "APP_RUNTIME_ROOT_PATH", "RUNTIME_ROOT_PATH")
	applyStringEnv(&AppSetting.LogLevel, "APP_LOG_LEVEL", "LOG_LEVEL")
	applyStringEnv(&AppSetting.InitData, "APP_INIT_DATA", "INIT_DATA")

	applyStringEnv(&ServerSetting.RunMode, "SERVER_RUN_MODE", "RUN_MODE")
	applyIntEnv(&ServerSetting.HttpPort, "SERVER_HTTP_PORT", "HTTP_PORT")
	applyDurationSecondsEnv(&ServerSetting.ReadTimeout, "SERVER_READ_TIMEOUT", "READ_TIMEOUT")
	applyDurationSecondsEnv(&ServerSetting.WriteTimeout, "SERVER_WRITE_TIMEOUT", "WRITE_TIMEOUT")
	applyStringEnv(&ServerSetting.EmbedHtml, "SERVER_EMBED_HTML", "EMBED_HTML")
	applyStringEnv(&ServerSetting.UrlPrefix, "SERVER_URL_PREFIX", "URL_PREFIX")

	applyStringEnv(&DatabaseSetting.Type, "DATABASE_TYPE", "DB_TYPE")
	applyStringEnv(&DatabaseSetting.User, "DATABASE_USER", "DB_USER", "MYSQL_USER")
	applyStringEnv(&DatabaseSetting.Password, "DATABASE_PASSWORD", "DB_PASSWORD", "MYSQL_PASSWORD")
	applyStringEnv(&DatabaseSetting.Host, "DATABASE_HOST", "DB_HOST", "MYSQL_HOST")
	applyIntEnv(&DatabaseSetting.Port, "DATABASE_PORT", "DB_PORT", "MYSQL_PORT")
	applyStringEnv(&DatabaseSetting.Name, "DATABASE_NAME", "DB_NAME", "MYSQL_DB")
	applyStringEnv(&DatabaseSetting.TablePrefix, "DATABASE_TABLE_PREFIX", "DB_TABLE_PREFIX", "MYSQL_TABLE_PREFIX")
	applyStringEnv(&DatabaseSetting.SqlDebug, "DATABASE_SQL_DEBUG", "SQL_DEBUG")
	applyStringEnv(&DatabaseSetting.Ssl, "DATABASE_SSL", "SSL")
}

func loadConfigFromEnv() {
	initDefaultSettings()
	applyEnvOverrides()
	printOptionValue()
}
