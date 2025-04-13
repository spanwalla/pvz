package repository

import "errors"

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrNoRowsDeleted = errors.New("no rows deleted")
)
