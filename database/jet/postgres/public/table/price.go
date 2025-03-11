//
// Code generated by go-jet DO NOT EDIT.
//
// WARNING: Changes to this file may cause incorrect behavior
// and will be lost if the code is regenerated
//

package table

import (
	"github.com/go-jet/jet/v2/postgres"
)

var Price = newPriceTable("public", "price", "")

type priceTable struct {
	postgres.Table

	// Columns
	ID           postgres.ColumnInteger
	Amount       postgres.ColumnFloat
	CurrencyCode postgres.ColumnString
	ProductID    postgres.ColumnInteger
	StoreID      postgres.ColumnInteger
	BranchID     postgres.ColumnInteger
	StockID      postgres.ColumnInteger
	CreatedByID  postgres.ColumnInteger
	UpdatedByID  postgres.ColumnInteger
	CreatedAt    postgres.ColumnTimestampz
	UpdatedAt    postgres.ColumnTimestampz

	AllColumns     postgres.ColumnList
	MutableColumns postgres.ColumnList
}

type PriceTable struct {
	priceTable

	EXCLUDED priceTable
}

// AS creates new PriceTable with assigned alias
func (a PriceTable) AS(alias string) *PriceTable {
	return newPriceTable(a.SchemaName(), a.TableName(), alias)
}

// Schema creates new PriceTable with assigned schema name
func (a PriceTable) FromSchema(schemaName string) *PriceTable {
	return newPriceTable(schemaName, a.TableName(), a.Alias())
}

// WithPrefix creates new PriceTable with assigned table prefix
func (a PriceTable) WithPrefix(prefix string) *PriceTable {
	return newPriceTable(a.SchemaName(), prefix+a.TableName(), a.TableName())
}

// WithSuffix creates new PriceTable with assigned table suffix
func (a PriceTable) WithSuffix(suffix string) *PriceTable {
	return newPriceTable(a.SchemaName(), a.TableName()+suffix, a.TableName())
}

func newPriceTable(schemaName, tableName, alias string) *PriceTable {
	return &PriceTable{
		priceTable: newPriceTableImpl(schemaName, tableName, alias),
		EXCLUDED:   newPriceTableImpl("", "excluded", ""),
	}
}

func newPriceTableImpl(schemaName, tableName, alias string) priceTable {
	var (
		IDColumn           = postgres.IntegerColumn("id")
		AmountColumn       = postgres.FloatColumn("amount")
		CurrencyCodeColumn = postgres.StringColumn("currency_code")
		ProductIDColumn    = postgres.IntegerColumn("product_id")
		StoreIDColumn      = postgres.IntegerColumn("store_id")
		BranchIDColumn     = postgres.IntegerColumn("branch_id")
		StockIDColumn      = postgres.IntegerColumn("stock_id")
		CreatedByIDColumn  = postgres.IntegerColumn("created_by_id")
		UpdatedByIDColumn  = postgres.IntegerColumn("updated_by_id")
		CreatedAtColumn    = postgres.TimestampzColumn("created_at")
		UpdatedAtColumn    = postgres.TimestampzColumn("updated_at")
		allColumns         = postgres.ColumnList{IDColumn, AmountColumn, CurrencyCodeColumn, ProductIDColumn, StoreIDColumn, BranchIDColumn, StockIDColumn, CreatedByIDColumn, UpdatedByIDColumn, CreatedAtColumn, UpdatedAtColumn}
		mutableColumns     = postgres.ColumnList{AmountColumn, CurrencyCodeColumn, ProductIDColumn, StoreIDColumn, BranchIDColumn, StockIDColumn, CreatedByIDColumn, UpdatedByIDColumn, CreatedAtColumn, UpdatedAtColumn}
	)

	return priceTable{
		Table: postgres.NewTable(schemaName, tableName, alias, allColumns...),

		//Columns
		ID:           IDColumn,
		Amount:       AmountColumn,
		CurrencyCode: CurrencyCodeColumn,
		ProductID:    ProductIDColumn,
		StoreID:      StoreIDColumn,
		BranchID:     BranchIDColumn,
		StockID:      StockIDColumn,
		CreatedByID:  CreatedByIDColumn,
		UpdatedByID:  UpdatedByIDColumn,
		CreatedAt:    CreatedAtColumn,
		UpdatedAt:    UpdatedAtColumn,

		AllColumns:     allColumns,
		MutableColumns: mutableColumns,
	}
}
