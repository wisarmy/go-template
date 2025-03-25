package errcode

// General status codes
const (
	Ok      = "ok"
	Unknown = "unknown.error"
)

// User service error codes
const (
	UserNotFound      = "user.not_found"
	UserUnauthorized  = "user.unauthorized"
	UserRegisterError = "user.register.error"
	UserLoginError    = "user.login.error"
	UserDisabled      = "user.disabled"
)

// Authentication related error codes
const (
	AuthTokenInvalid = "auth.token.invalid"
	AuthTokenExpired = "auth.token.expired"
	AuthAccessDenied = "auth.access.denied"
)

// Role related error codes
const (
	RoleNotFound = "role.not_found"
	RoleInUse    = "role.in_use"
)

// Parameter validation error codes
const (
	InvalidParams = "invalid.params"
)

// System level error codes
const (
	ServerError  = "server.error"
	DBError      = "db.error"
	NetworkError = "network.error"
)

// Resource error codes
const (
	ResourceNotFound  = "resource.not_found"
	ResourceForbidden = "resource.forbidden"
)

// Message mapping for error codes
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

// Error represents an error with a code and message
type Error struct {
	Code    string `json:"code"`    // Error code
	Message string `json:"message"` // Error message
}

// New creates a new error with the specified code
func New(code string) *Error {
	return &Error{
		Code:    code,
		Message: GetMessage(code),
	}
}

// WithMessage returns an error with a custom message
func (e *Error) WithMessage(message string) *Error {
	e.Message = message
	return e
}

// Error implements the error interface
func (e *Error) Error() string {
	return e.Message + " (Code: " + e.Code + ")"
}
