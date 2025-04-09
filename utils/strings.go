package utils

import (
	"fmt"
	"strconv"
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

func PostgresArrayToStrArray(p_array string) []string {
	return strings.Split(strings.Trim(p_array, "{}"), ",")
}

func PostgresArrayToIntArray(p_array string) []int {
	var err error
	str_array := strings.Split(strings.Trim(p_array, "{}"), ",")
	int_array := make([]int, len(str_array))
	for i, v := range str_array {
		int_array[i], err = strconv.Atoi(v)
		if err != nil {
			panic("could not process value as int")
		}
	}
	return int_array
}
