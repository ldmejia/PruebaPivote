package repository

import (
	"errors"

	mssql "github.com/microsoft/go-mssqldb"
)

func isDuplicateKey(err error) bool {
	var e mssql.Error
	if errors.As(err, &e) {
		return e.Number == 2627 || e.Number == 2601
	}
	return false
}
