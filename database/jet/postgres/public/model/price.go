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

type Price struct {
	ID            int64 `sql:"primary_key"`
	Amount        float64
	CurrencyCode  string
	ProductID     int64
	StoreID       int64
	BranchID      int64
	StockID       int64
	CreatedByID   *int64
	UpdatedByID   *int64
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Sale          bool
	OriginalPrice *float64
	Condition     *string
	UnitType      string
	ImageID       *string
	ExpiresAt     *time.Time
}
