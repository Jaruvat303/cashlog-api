package domain

import "errors"

// ประกาศตัวแปร Error ที่อาจเกิดขึ้นได้ในระบบธุรกิจของเรา
var (
	ErrNotFound         = errors.New("data not found")
	ErrDuplicateRequest = errors.New("duplicate request detected")
	ErrInvalidInput     = errors.New("invalid input parameters")
	ErrInternalDB       = errors.New("internal database error")
	ErrContextCanceled  = errors.New("request canceled by user")
	ErrTimeout          = errors.New("database query timeout exceeded")
)
