package oauth

import "errors"

var (
	ErrInvalidToken      = errors.New("invalid token")
	ErrInsufficientScope = errors.New("insufficient scope")
	ErrSubjectNotFound   = errors.New("subject not found")
	ErrConfig            = errors.New("oauth config error")
)

func IsInvalidToken(err error) bool {
	return errors.Is(err, ErrInvalidToken)
}
