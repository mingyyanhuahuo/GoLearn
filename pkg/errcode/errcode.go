package errcode

type BizError struct {
	HttpStatus int
	Code       int
	Message    string
}

func (e *BizError) Error() string {
	return e.Message
}
func New(httpStatus int, code int, msg string) *BizError {
	return &BizError{
		HttpStatus: httpStatus,
		Code:       code,
		Message:    msg,
	}
}

var (
	ErrBadRequest          = &BizError{400, 10000, "参数不合法"}
	ErrUnauthorized        = &BizError{401, 10001, "未登录或登录已过期"}
	ErrForbidden           = &BizError{403, 10002, "没有权限"}
	ErrNotFoundPost        = &BizError{404, 10003, "帖子不存在"}
	ErrNotFoundUser        = &BizError{404, 10004, "用户不存在"}
	ErrNotFoundComment     = &BizError{404, 10005, "评论不存在"}
	ErrDatabase            = &BizError{500, 10006, "数据库错误"}
	ErrExistingUser        = &BizError{409, 10007, "用户已存在"}
	ErrInternalServerError = &BizError{500, 10008, "服务器内部错误"}
	ErrWrongCredentials    = &BizError{401, 10009, "账号或密码错误"}
	ErrRateLimit           = &BizError{429, 10012, "请求过于频繁，请稍后再试"}
	ErrDraftNotFound       = &BizError{404, 10010, "待确认草稿不存在或已过期"}
	ErrDraftInvalid        = &BizError{400, 10011, "草稿确认无效"}
)
