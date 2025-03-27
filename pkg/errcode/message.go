package errcode

var messages = map[string]string{
	Ok:      "操作成功",
	Unknown: "未知错误",

	UserNotFound:      "用户不存在",
	UserUnauthorized:  "用户未授权",
	UserRegisterError: "用户注册失败",
	UserLoginError:    "用户登录失败",
	UserDisabled:      "用户已被禁用",

	AuthTokenInvalid: "无效的认证令牌",
	AuthTokenExpired: "认证令牌已过期",
	AuthAccessDenied: "拒绝访问",

	InvalidParams: "无效的参数",

	ResourceNotFound:  "资源不存在",
	ResourceForbidden: "禁止访问此资源",

	ServerError:  "服务器内部错误",
	DBError:      "数据库操作失败",
	NetworkError: "网络通信错误",

	RoleNotFound: "角色不存在",
	RoleInUse:    "角色正在使用中，无法删除",
}

// GetMessage returns the message for a given error code
func GetMessage(code string) string {
	if msg, ok := messages[code]; ok {
		return msg
	}
	return messages[Unknown]
}
