package utils

import (
	"fmt"
	"strings"

	"github.com/go-jet/jet/v2/postgres"
)

func BuildFullTableName(col postgres.Column) string {
	return fmt.Sprintf(`%s.%s`, col.TableName(), col.Name())
}

func ToPostgresArray[T comparable](arr []T) string {
	str_arr := make([]string, len(arr))
	for i, val := range arr {
		str_arr[i] = fmt.Sprintf("%v", val)
	}
	return fmt.Sprintf("{%s}", strings.Join(str_arr, ","))
}
