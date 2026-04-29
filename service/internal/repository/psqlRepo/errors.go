package psqlrepo

import "errors"

var (
	ErrNotFound         = errors.New("resource not found")
	ErrDuplicateEmail   = errors.New("email already exists")
	ErrDuplicatePhone   = errors.New("phone already exists")
	ErrNoFieldsToUpdate = errors.New("no fields to update")
	ErrUpdateFailed     = errors.New("failed to update resource")
)
