//
// Code generated by go-jet DO NOT EDIT.
//
// WARNING: Changes to this file may cause incorrect behavior
// and will be lost if the code is regenerated
//

package enum

import "github.com/go-jet/jet/v2/postgres"

var UserAuthPlatformType = &struct {
	Internal postgres.StringExpression
	Apple    postgres.StringExpression
	Google   postgres.StringExpression
}{
	Internal: postgres.NewEnumValue("INTERNAL"),
	Apple:    postgres.NewEnumValue("APPLE"),
	Google:   postgres.NewEnumValue("GOOGLE"),
}
