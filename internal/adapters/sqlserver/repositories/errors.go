package repository

import (
	"errors"

	mssql "github.com/microsoft/go-mssqldb"
)

// sin esto el driver manda nvarchar y sql server no usa el indice
func vc(s string) mssql.VarChar {
	return mssql.VarChar(s)
}

func isDeadlock(err error) bool {
	var e mssql.Error
	return errors.As(err, &e) && e.Number == 1205
}

func isDuplicateKey(err error) bool {
	var e mssql.Error
	if errors.As(err, &e) {
		return e.Number == 2627 || e.Number == 2601
	}
	return false
}
