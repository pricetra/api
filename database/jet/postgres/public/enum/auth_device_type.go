//
// Code generated by go-jet DO NOT EDIT.
//
// WARNING: Changes to this file may cause incorrect behavior
// and will be lost if the code is regenerated
//

package enum

import "github.com/go-jet/jet/v2/postgres"

var AuthDeviceType = &struct {
	Ios     postgres.StringExpression
	Android postgres.StringExpression
	Web     postgres.StringExpression
	Other   postgres.StringExpression
	Unknown postgres.StringExpression
}{
	Ios:     postgres.NewEnumValue("ios"),
	Android: postgres.NewEnumValue("android"),
	Web:     postgres.NewEnumValue("web"),
	Other:   postgres.NewEnumValue("other"),
	Unknown: postgres.NewEnumValue("unknown"),
}
