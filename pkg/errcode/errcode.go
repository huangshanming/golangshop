package errcode

const (
	OK = 0

	ErrBadRequest   = 40000
	ErrUnauthorized = 40100
	ErrForbidden    = 40300
	ErrNotFound     = 40400
	ErrInternal     = 50000

	ErrTenantMissing = 40001
	ErrTokenInvalid  = 40101
	ErrTokenExpired  = 40102

	ErrUserNotFound    = 40401
	ErrProductNotFound = 40402
	ErrOrderNotFound   = 40403
	ErrStockNotEnough  = 40010
	ErrOrderDuplicate  = 40011
	ErrPayFailed       = 40020
)
