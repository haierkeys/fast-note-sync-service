package global

import (
	"github.com/haierkeys/fast-note-sync-service/pkg/validator"

	ut "github.com/go-playground/universal-translator"
)

var (
	Validator *validator.CustomValidator
	Ut        *ut.UniversalTranslator
)
