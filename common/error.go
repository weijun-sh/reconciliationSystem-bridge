package common

import (
	"errors"
)

var (
	ErrAddressNull    = errors.New("Address is Null")
	ErrAddressInValid = errors.New("Address is InValid")
)

var (
	ErrNoValueObtaind = errors.New("No Value Obtaind")
)
