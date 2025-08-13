package util

import (
	"fmt"
)

func GetDBConnectionString(host, user, password, dbName string, port int) string {
	// Note that we don't assume host can be IPv6 format including colon(:) in itself
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&collation=utf8mb4_bin&parseTime=True&loc=Local&interpolateParams=true",
		user, password, host, port, dbName)
}
