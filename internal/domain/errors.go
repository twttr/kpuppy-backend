package domain

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserBanned         = errors.New("user is banned")
	ErrUsernameEmpty      = errors.New("username cannot be empty")
	ErrContentNotFound    = errors.New("content not found")
	ErrCommentNotFound = errors.New("comment not found")
	ErrCommentDeleted  = errors.New("comment has been deleted")
	ErrNotCommentOwner = errors.New("user is not the comment owner")
	ErrCommentTooLong     = errors.New("comment exceeds 1000 characters")
	ErrCommentEmpty       = errors.New("comment text cannot be empty")
	ErrUserNotProvisioned = errors.New("user not provisioned")
	ErrRateLimitExceeded  = errors.New("rate limit exceeded")
)

type ErrorCode string

const (
	CodeUserNotProvisioned ErrorCode = "USER_NOT_PROVISIONED"
	CodeContentNotFound    ErrorCode = "CONTENT_NOT_FOUND"
	CodeCommentNotFound    ErrorCode = "COMMENT_NOT_FOUND"
	CodeValidationError    ErrorCode = "VALIDATION_ERROR"
	CodeForbidden          ErrorCode = "FORBIDDEN"
	CodeUserBanned         ErrorCode = "USER_BANNED"
	CodeRateLimitExceeded  ErrorCode = "RATE_LIMIT_EXCEEDED"
)

type APIError struct {
	Error string    `json:"error"`
	Code  ErrorCode `json:"code"`
}

func NewAPIError(err error, code ErrorCode) APIError {
	return APIError{
		Error: err.Error(),
		Code:  code,
	}
}
