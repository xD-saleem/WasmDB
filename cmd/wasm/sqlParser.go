package main

import (
	"fmt"
	"strings"

	"github.com/google/btree"
	"github.com/google/uuid"
	"github.com/xwb1989/sqlparser"
)

type dynamicValue map[string]any

func createDynamicStruct(columnNames, values []string) dynamicValue {
	if len(columnNames) != len(values) {
		panic("Column names and values must have the same length")
	}

	data := make(map[string]any)

	for i := 0; i < len(columnNames); i++ {
		columnName := strings.ReplaceAll(columnNames[i], "'", "")
		value := strings.ReplaceAll(values[i], "'", "")
		data[columnName] = value
	}

	return data
}

func getTableName(stmt sqlparser.Statement) (string, error) {
	switch s := stmt.(type) {
	case *sqlparser.Select:
		// Iterate over the FROM clause
		if s.From != nil && s.From[0] != nil {
			tableName := sqlparser.String(s.From[0])
			return tableName, nil
		}
		return "", fmt.Errorf("No table name found")
	case *sqlparser.Insert:
		tableName := sqlparser.String(s.Table)
		return tableName, nil
	default:
		return "", fmt.Errorf("Unsupported statement type")
	}
}

func getColumnNames(stmt sqlparser.Statement) ([]string, error) {
	switch s := stmt.(type) {
	case *sqlparser.Select:
		var columnNames []string
		// Iterate over the SELECT clause
		for _, expr := range s.SelectExprs {
			switch e := expr.(type) {
			case *sqlparser.StarExpr:
				// Handle the case where * is used to select all columns
				columnNames = append(columnNames, "*")
			case *sqlparser.AliasedExpr:
				// Extract column name
				colName := e.Expr.(*sqlparser.ColName).Name.String()
				columnNames = append(columnNames, colName)
			default:
				return nil, fmt.Errorf("Unsupported expression type in SELECT clause")
			}
		}
		return columnNames, nil
	case *sqlparser.Insert:
		var columnNames []string
		// Iterate over the column getColumnNames
		for _, col := range s.Columns {
			colName := sqlparser.String(col)
			columnNames = append(columnNames, colName)
		}
		return columnNames, nil
	default:
		return nil, fmt.Errorf("Unsupported statement type")
	}
}

// ############# SELECT #############

func buildSelectAction(stmt sqlparser.Statement) *Action {
	// Handle SELECT statement
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
		// Handle error
		fmt.Println("Error getting table name:", err)
		return nil
	}

	columnNames, err := getColumnNames(s)
	if err != nil {
		// Handle error
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

func (ds *DatabaseService) buildAction(stmt sqlparser.Statement) *Action {
	switch stmt.(type) {
	case *sqlparser.Select:
		return buildSelectAction(stmt)
	case *sqlparser.Insert:
		return buildInsertAction(stmt)
	default:
		fmt.Println("Unsupported statement type")
		return nil
	}
}

func (ds *DatabaseService) execAction(action *Action) {
	switch action.Type {
	case "SELECT":
		if action.ColumnNames[0] == "*" {
			ds.Btree.Ascend(func(item btree.Item) bool {
				node := item.(Node)
				fmt.Println(node.val)
				return true
			})
			return
		}

		var foundMap = make(map[string]int)
		var foundNode []Node

		valueToFind := action.Conditions
		sv := getStructValues(valueToFind[0])
		mapSize := len(sv)

		// Iterate over the tree
		ds.Btree.Ascend(func(item btree.Item) bool {
			node := item.(Node)
			for k, v := range sv {
				if v == node.val[k] {
					foundMap[k] = foundMap[k] + 1
					foundNode = append(foundNode, node)
				}
			}
			return true
		})

		// Check if any nodes were found
		if len(foundMap) == mapSize {
			fmt.Println("Found nodes:", foundNode)
		} else {
			fmt.Println("No nodes found with the specified value.")
		}

	case "INSERT":
		// Inserting into the tree
		s := createDynamicStruct(action.ColumnNames, action.Values)
		ds.Btree.ReplaceOrInsert(Node{Id: int(uuid.New().ID()), val: s})
	case "DELETE":
	case "UPDATE":
	default:
		fmt.Println("Unsupported action type")
	}
}

func getStructValues(s string) dynamicValue {
	columns := []string{}
	values := []string{}

	s = strings.ReplaceAll(s, "and", "")
	s = strings.ReplaceAll(s, "'", "")
	s = strings.ReplaceAll(s, ";", "")

	d := strings.Split(s, "  ")
	for _, v := range d {
		s := strings.Split(v, " = ")
		columns = append(columns, s[0])
		values = append(values, s[1])
	}

	return createDynamicStruct(columns, values)
}
