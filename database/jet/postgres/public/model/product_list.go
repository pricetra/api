//
// Code generated by go-jet DO NOT EDIT.
//
// WARNING: Changes to this file may cause incorrect behavior
// and will be lost if the code is regenerated
//

package model

import (
	"time"
)

type ProductList struct {
	ID        int64 `sql:"primary_key"`
	UserID    int64
	ListID    int64
	ProductID int64
	StockID   *int64
	CreatedAt time.Time
}
