package errcode

// General codes
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

// Authentication error codes
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
