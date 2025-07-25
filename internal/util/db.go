package util

import "fmt"

func GetDBConnectionString(host, user, password, dbName string, port int) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&collation=utf8mb4_bin&parseTime=True&loc=Local&interpolateParams=true",
		user, password, host, port, dbName)
}
