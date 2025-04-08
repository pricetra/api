package utils

import (
	"fmt"

	"github.com/go-jet/jet/v2/postgres"
)

func BuildFullTableName(col postgres.Column) string {
	return fmt.Sprintf(`%s.%s`, col.TableName(), col.Name())
}
