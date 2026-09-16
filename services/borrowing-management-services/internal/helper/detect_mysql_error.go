package helper

import (
	"errors"
	"strings"

	"github.com/go-sql-driver/mysql"
)

const mysqlDuplicateEntryErrorCode = 1062

func IsDuplicateEntryError(err error, keyName string) bool {

	var mysqlError *mysql.MySQLError

	if !errors.As(err, &mysqlError) {
		return false
	}

	if mysqlError.Number != mysqlDuplicateEntryErrorCode {
		return false
	}

	return strings.Contains(mysqlError.Message, keyName)

}
