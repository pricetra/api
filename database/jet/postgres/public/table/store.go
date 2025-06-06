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

var Store = newStoreTable("public", "store", "")

type storeTable struct {
	postgres.Table

	// Columns
	ID          postgres.ColumnInteger
	Name        postgres.ColumnString
	Logo        postgres.ColumnString
	Website     postgres.ColumnString
	CreatedByID postgres.ColumnInteger
	UpdatedByID postgres.ColumnInteger
	CreatedAt   postgres.ColumnTimestampz
	UpdatedAt   postgres.ColumnTimestampz

	AllColumns     postgres.ColumnList
	MutableColumns postgres.ColumnList
}

type StoreTable struct {
	storeTable

	EXCLUDED storeTable
}

// AS creates new StoreTable with assigned alias
func (a StoreTable) AS(alias string) *StoreTable {
	return newStoreTable(a.SchemaName(), a.TableName(), alias)
}

// Schema creates new StoreTable with assigned schema name
func (a StoreTable) FromSchema(schemaName string) *StoreTable {
	return newStoreTable(schemaName, a.TableName(), a.Alias())
}

// WithPrefix creates new StoreTable with assigned table prefix
func (a StoreTable) WithPrefix(prefix string) *StoreTable {
	return newStoreTable(a.SchemaName(), prefix+a.TableName(), a.TableName())
}

// WithSuffix creates new StoreTable with assigned table suffix
func (a StoreTable) WithSuffix(suffix string) *StoreTable {
	return newStoreTable(a.SchemaName(), a.TableName()+suffix, a.TableName())
}

func newStoreTable(schemaName, tableName, alias string) *StoreTable {
	return &StoreTable{
		storeTable: newStoreTableImpl(schemaName, tableName, alias),
		EXCLUDED:   newStoreTableImpl("", "excluded", ""),
	}
}

func newStoreTableImpl(schemaName, tableName, alias string) storeTable {
	var (
		IDColumn          = postgres.IntegerColumn("id")
		NameColumn        = postgres.StringColumn("name")
		LogoColumn        = postgres.StringColumn("logo")
		WebsiteColumn     = postgres.StringColumn("website")
		CreatedByIDColumn = postgres.IntegerColumn("created_by_id")
		UpdatedByIDColumn = postgres.IntegerColumn("updated_by_id")
		CreatedAtColumn   = postgres.TimestampzColumn("created_at")
		UpdatedAtColumn   = postgres.TimestampzColumn("updated_at")
		allColumns        = postgres.ColumnList{IDColumn, NameColumn, LogoColumn, WebsiteColumn, CreatedByIDColumn, UpdatedByIDColumn, CreatedAtColumn, UpdatedAtColumn}
		mutableColumns    = postgres.ColumnList{NameColumn, LogoColumn, WebsiteColumn, CreatedByIDColumn, UpdatedByIDColumn, CreatedAtColumn, UpdatedAtColumn}
	)

	return storeTable{
		Table: postgres.NewTable(schemaName, tableName, alias, allColumns...),

		//Columns
		ID:          IDColumn,
		Name:        NameColumn,
		Logo:        LogoColumn,
		Website:     WebsiteColumn,
		CreatedByID: CreatedByIDColumn,
		UpdatedByID: UpdatedByIDColumn,
		CreatedAt:   CreatedAtColumn,
		UpdatedAt:   UpdatedAtColumn,

		AllColumns:     allColumns,
		MutableColumns: mutableColumns,
	}
}
