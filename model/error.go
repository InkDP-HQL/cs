package model

import "errors"

// 常用错误
var (
	ErrAccountIsEmpty          = errors.New("account is empty")
	ErrPasswordIsEmpty         = errors.New("password is empty")
	ErrPhoneIsEmpty            = errors.New("phone is empty")
	ErrEmailIsEmpty            = errors.New("email is empty")
	ErrAccountNotMatchPassword = errors.New("account and password don't match")
	ErrRepeatingAccountName    = errors.New("repeating account name")

	ErrTagName = errors.New("repeating tag name")

	ErrInvalidProgressValue = errors.New("invalid progress value")

	ErrNotFound               = errors.New("find no result")
	ErrNeedsPtr               = errors.New("need a ptr param")
	ErrIndexNotFoundException = errors.New("index not fund")
)
