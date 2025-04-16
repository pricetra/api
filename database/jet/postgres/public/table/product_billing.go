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

var ProductBilling = newProductBillingTable("public", "product_billing", "")

type productBillingTable struct {
	postgres.Table

	// Columns
	ID              postgres.ColumnInteger
	ProductID       postgres.ColumnInteger
	UserID          postgres.ColumnInteger
	CreatedAt       postgres.ColumnTimestampz
	Rate            postgres.ColumnFloat
	BillingRateType postgres.ColumnString
	NewData         postgres.ColumnString
	OldData         postgres.ColumnString

	AllColumns     postgres.ColumnList
	MutableColumns postgres.ColumnList
}

type ProductBillingTable struct {
	productBillingTable

	EXCLUDED productBillingTable
}

// AS creates new ProductBillingTable with assigned alias
func (a ProductBillingTable) AS(alias string) *ProductBillingTable {
	return newProductBillingTable(a.SchemaName(), a.TableName(), alias)
}

// Schema creates new ProductBillingTable with assigned schema name
func (a ProductBillingTable) FromSchema(schemaName string) *ProductBillingTable {
	return newProductBillingTable(schemaName, a.TableName(), a.Alias())
}

// WithPrefix creates new ProductBillingTable with assigned table prefix
func (a ProductBillingTable) WithPrefix(prefix string) *ProductBillingTable {
	return newProductBillingTable(a.SchemaName(), prefix+a.TableName(), a.TableName())
}

// WithSuffix creates new ProductBillingTable with assigned table suffix
func (a ProductBillingTable) WithSuffix(suffix string) *ProductBillingTable {
	return newProductBillingTable(a.SchemaName(), a.TableName()+suffix, a.TableName())
}

func newProductBillingTable(schemaName, tableName, alias string) *ProductBillingTable {
	return &ProductBillingTable{
		productBillingTable: newProductBillingTableImpl(schemaName, tableName, alias),
		EXCLUDED:            newProductBillingTableImpl("", "excluded", ""),
	}
}

func newProductBillingTableImpl(schemaName, tableName, alias string) productBillingTable {
	var (
		IDColumn              = postgres.IntegerColumn("id")
		ProductIDColumn       = postgres.IntegerColumn("product_id")
		UserIDColumn          = postgres.IntegerColumn("user_id")
		CreatedAtColumn       = postgres.TimestampzColumn("created_at")
		RateColumn            = postgres.FloatColumn("rate")
		BillingRateTypeColumn = postgres.StringColumn("billing_rate_type")
		NewDataColumn         = postgres.StringColumn("new_data")
		OldDataColumn         = postgres.StringColumn("old_data")
		allColumns            = postgres.ColumnList{IDColumn, ProductIDColumn, UserIDColumn, CreatedAtColumn, RateColumn, BillingRateTypeColumn, NewDataColumn, OldDataColumn}
		mutableColumns        = postgres.ColumnList{ProductIDColumn, UserIDColumn, CreatedAtColumn, RateColumn, BillingRateTypeColumn, NewDataColumn, OldDataColumn}
	)

	return productBillingTable{
		Table: postgres.NewTable(schemaName, tableName, alias, allColumns...),

		//Columns
		ID:              IDColumn,
		ProductID:       ProductIDColumn,
		UserID:          UserIDColumn,
		CreatedAt:       CreatedAtColumn,
		Rate:            RateColumn,
		BillingRateType: BillingRateTypeColumn,
		NewData:         NewDataColumn,
		OldData:         OldDataColumn,

		AllColumns:     allColumns,
		MutableColumns: mutableColumns,
	}
}
