package e

var MsgFlags = map[int]string{
	SUCCESS:                        "ok",
	ERROR:                          "fail",
	INVALID_PARAMS:                 "请求参数错误",
	ERROR_AUTH_CHECK_TOKEN_FAIL:    "Token鉴权失败",
	ERROR_AUTH_CHECK_TOKEN_TIMEOUT: "Token已超时",
	ERROR_AUTH_TOKEN:               "Token生成失败",
	ERROR_AUTH_NO_TOKEN:            "Token缺失",
	ERROR_AUTH:                     "Token错误",
	ERROR_AUTH_FORBIDDEN:           "无权限访问",

	// MQ 数据源相关
	ERROR_GET_SOURCE_FAIL:    "获取数据源失败",
	ERROR_ADD_SOURCE_FAIL:    "新增数据源失败",
	ERROR_EDIT_SOURCE_FAIL:   "编辑数据源失败",
	ERROR_DELETE_SOURCE_FAIL: "删除数据源失败",

	// 订阅相关
	ERROR_GET_SUBSCRIPTION_FAIL:    "获取订阅失败",
	ERROR_ADD_SUBSCRIPTION_FAIL:    "新增订阅失败",
	ERROR_EDIT_SUBSCRIPTION_FAIL:   "编辑订阅失败",
	ERROR_DELETE_SUBSCRIPTION_FAIL: "删除订阅失败",
	ERROR_START_SUBSCRIPTION_FAIL:  "启动订阅失败",
	ERROR_STOP_SUBSCRIPTION_FAIL:   "停止订阅失败",

	// 消费日志相关
	ERROR_GET_CONSUME_LOG_FAIL: "获取消费日志失败",
}

// GetMsg get error information based on Code
func GetMsg(code int) string {
	msg, ok := MsgFlags[code]
	if ok {
		return msg
	}

	return MsgFlags[ERROR]
}
