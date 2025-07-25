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

var AiPromptResponse = newAiPromptResponseTable("public", "ai_prompt_response", "")

type aiPromptResponseTable struct {
	postgres.Table

	// Columns
	ID        postgres.ColumnInteger
	Type      postgres.ColumnString
	Request   postgres.ColumnString
	Response  postgres.ColumnString
	UserID    postgres.ColumnInteger
	CreatedAt postgres.ColumnTimestampz

	AllColumns     postgres.ColumnList
	MutableColumns postgres.ColumnList
}

type AiPromptResponseTable struct {
	aiPromptResponseTable

	EXCLUDED aiPromptResponseTable
}

// AS creates new AiPromptResponseTable with assigned alias
func (a AiPromptResponseTable) AS(alias string) *AiPromptResponseTable {
	return newAiPromptResponseTable(a.SchemaName(), a.TableName(), alias)
}

// Schema creates new AiPromptResponseTable with assigned schema name
func (a AiPromptResponseTable) FromSchema(schemaName string) *AiPromptResponseTable {
	return newAiPromptResponseTable(schemaName, a.TableName(), a.Alias())
}

// WithPrefix creates new AiPromptResponseTable with assigned table prefix
func (a AiPromptResponseTable) WithPrefix(prefix string) *AiPromptResponseTable {
	return newAiPromptResponseTable(a.SchemaName(), prefix+a.TableName(), a.TableName())
}

// WithSuffix creates new AiPromptResponseTable with assigned table suffix
func (a AiPromptResponseTable) WithSuffix(suffix string) *AiPromptResponseTable {
	return newAiPromptResponseTable(a.SchemaName(), a.TableName()+suffix, a.TableName())
}

func newAiPromptResponseTable(schemaName, tableName, alias string) *AiPromptResponseTable {
	return &AiPromptResponseTable{
		aiPromptResponseTable: newAiPromptResponseTableImpl(schemaName, tableName, alias),
		EXCLUDED:              newAiPromptResponseTableImpl("", "excluded", ""),
	}
}

func newAiPromptResponseTableImpl(schemaName, tableName, alias string) aiPromptResponseTable {
	var (
		IDColumn        = postgres.IntegerColumn("id")
		TypeColumn      = postgres.StringColumn("type")
		RequestColumn   = postgres.StringColumn("request")
		ResponseColumn  = postgres.StringColumn("response")
		UserIDColumn    = postgres.IntegerColumn("user_id")
		CreatedAtColumn = postgres.TimestampzColumn("created_at")
		allColumns      = postgres.ColumnList{IDColumn, TypeColumn, RequestColumn, ResponseColumn, UserIDColumn, CreatedAtColumn}
		mutableColumns  = postgres.ColumnList{TypeColumn, RequestColumn, ResponseColumn, UserIDColumn, CreatedAtColumn}
	)

	return aiPromptResponseTable{
		Table: postgres.NewTable(schemaName, tableName, alias, allColumns...),

		//Columns
		ID:        IDColumn,
		Type:      TypeColumn,
		Request:   RequestColumn,
		Response:  ResponseColumn,
		UserID:    UserIDColumn,
		CreatedAt: CreatedAtColumn,

		AllColumns:     allColumns,
		MutableColumns: mutableColumns,
	}
}
