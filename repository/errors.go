package repository

import (
	"errors"

	"github.com/bldsoft/gost/utils"
)

var (
	ErrAlreadyExists = errors.New("already exists")
	ErrNotFound      = utils.ErrObjectNotFound
)
