package main

import (
	"fmt"

	"github.com/xwb1989/sqlparser"
)

func buildSelectAction(stmt sqlparser.Statement) *Action {
	myAction := Action{}
	myAction.Type = "SELECT"

	// Extracting conditions
	s := stmt.(*sqlparser.Select)

	if s.Where != nil {
		condition := sqlparser.String(s.Where.Expr)
		myAction.Conditions = append(myAction.Conditions, condition)
	}

	// Extract Limit
	if s.Limit != nil {
		limit := sqlparser.String(s.Limit)
		myAction.Conditions = append(myAction.Conditions, limit)
	}

	tableName, err := getTableName(s)
	if err != nil {
		fmt.Println("Error getting table name:", err)
		return nil
	}

	columnNames, err := getColumnNames(s)
	if err != nil {
		fmt.Println("Error getting column names:", err)
		return nil
	}

	myAction.TableNames = append(myAction.TableNames, tableName)
	myAction.ColumnNames = append(myAction.ColumnNames, columnNames...)

	return &myAction
}

func buildInsertAction(stmt sqlparser.Statement) *Action {
	myAction := Action{}
	myAction.Type = "INSERT"

	// Extracting conditions
	s := stmt.(*sqlparser.Insert)
	tableName, err := getTableName(s)
	if err != nil {
		fmt.Println("Error getting table name: 2", err)
		return nil
	}

	columnNames, err := getColumnNames(s)
	if err != nil {
		fmt.Println("Error getting column names:", err)
		return nil
	}

	// Extracting values
	var values []string
	for _, row := range s.Rows.(sqlparser.Values) {
		for _, value := range row {
			values = append(values, sqlparser.String(value))
		}
	}
	myAction.Values = values

	myAction.TableNames = append(myAction.TableNames, tableName)
	myAction.ColumnNames = append(myAction.ColumnNames, columnNames...)

	return &myAction
}
