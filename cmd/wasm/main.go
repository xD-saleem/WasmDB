package main

import (
	"fmt"
	"strings"
	"syscall/js"

	"github.com/tidwall/btree"
	"github.com/xwb1989/sqlparser"
)

type Service[T interface{}] interface {
	SQL(sql *SQL) js.Func
}

type Database[T interface{}] struct {
	tree *btree.Map[string, string]
}

func NewDatabaseService[T any](tree *btree.Map[string, string]) Service[T] {
	return Database[T]{
		tree: tree,
	}
}

type SQL struct {
	key   string
	value string
}

type Customers struct {
	Name string
	Age  int
}

func (d Database[T]) SQL(sql *SQL) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) == 0 {
			return nil
		}
		sqlQuery := args[0]
		queryAtts := parseQuery(sqlQuery.String())

		d.execQuery(queryAtts)

		return nil
	})
}

func (d Database[T]) execQuery(queryAtts QueryAttributes) {
	if queryAtts.TableName == "" {
		fmt.Println("Table name is required")
		return
	} else {
		fmt.Println("Table name is", queryAtts.TableName)
	}

	switch strings.ToUpper(queryAtts.Action) {
	case "SELECT":
		fmt.Println("SELECT")
		if queryAtts.WhereClause != "" {
			fmt.Println("WHERE", queryAtts.WhereClause)
		}

		if queryAtts.Limit != "" {
			fmt.Println("LIMIT", queryAtts.Limit)
		}

		if queryAtts.SelectQuery == nil {
			fmt.Println("No columns provided")
		} else {
			fmt.Println("SELECT ", queryAtts.SelectQuery)
		}

		if queryAtts.TableName != "" {
			fmt.Println("FROM", queryAtts.TableName)
		}

		d.tree.Scan(func(k string, v string) bool {
			fmt.Printf("FOUND IN DB %+v\n", v)
			return true
		})

	case "INSERT":
		fmt.Println("INSERT")
		fmt.Println("INTO", queryAtts.TableName)
		fmt.Println("VALUES", queryAtts.Values)
		fmt.Println("COLUMNS", queryAtts.Columns)

		if queryAtts.Values == nil {
			fmt.Println("No values provided")
			return
		}

		c := Customers{}

		for i := 0; i < len(queryAtts.Columns); i++ {
			switch queryAtts.Columns[i] {
			// case "id":
			// 	c.Id = queryAtts.Values[i+1].(string)
			case "name":
				c.Name = queryAtts.Values[i+1].(string)
			case "age":
				c.Age = queryAtts.Values[i+1].(int)
			}
		}
		// id := uuid.New().String()

		d.tree.Set("1", c.Name)
		g, _ := d.tree.Get("1")
		fmt.Printf("%+v\n", g)

		d.tree.Scan(func(k string, v string) bool {
			fmt.Printf("%+v\n", v)
			return true
		})

	// case "UPDATE":
	// 	fmt.Println("UPDATE")
	// 	fmt.Println("SET", queryAtts.Values)
	// 	if queryAtts.Conditions != "" {
	// 		fmt.Println("WHERE", queryAtts.Conditions)
	// 	}
	// 	if queryAtts.TableName != "" {
	// 		fmt.Println("FROM", queryAtts.TableName)
	// 	}
	// 	if queryAtts.Values == nil {
	// 		fmt.Println("No values provided")
	// 		return
	// 	}
	//
	// 	if queryAtts.Conditions != "" {
	// 		fmt.Println("WHERE", queryAtts.Conditions)
	// 	}
	//
	// 	for i := 0; i < len(queryAtts.Values); i += 2 {
	// 		tree.Set(queryAtts.Values[i], queryAtts.Values[i+1])
	// 	}
	//
	// case "DELETE":
	// 	fmt.Println("DELETE")
	// 	if queryAtts.Conditions != "" {
	// 	}
	// 		fmt.Println("WHERE", queryAtts.Conditions)
	// 	if queryAtts.TableName != "" {
	// 		fmt.Println("FROM", queryAtts.TableName)
	// 	}

	default:
		fmt.Println("Unsupported SQL statement type")
	}
}

type QueryAttributes struct {
	Action    string
	TableName string
	Column    string
	Columns   []string
	Values    []any

	Conditions  string
	WhereClause string
	Limit       string
	SelectQuery []string
}

func parseQuery(sqlQuery string) QueryAttributes {
	stmt, err := sqlparser.Parse(sqlQuery)
	if err != nil {
		fmt.Println("Failed to parse SQL query", err)
	}

	queryAttributes := QueryAttributes{}

	switch stmt := stmt.(type) {
	case *sqlparser.Select:
		queryAttributes.Action = "SELECT"
		if stmt.From != nil && stmt.From[0] != nil {
			queryAttributes.TableName = sqlparser.String(stmt.From[0])

		}
		if len(stmt.SelectExprs) == 1 {
			if _, ok := stmt.SelectExprs[0].(*sqlparser.StarExpr); ok {
				queryAttributes.SelectQuery = append(queryAttributes.SelectQuery, "*")
			}
		} else {
			for _, expr := range stmt.SelectExprs {
				queryAttributes.SelectQuery = append(queryAttributes.SelectQuery, sqlparser.String(expr))
			}
		}
		if stmt.Where != nil {
			whereExpr := sqlparser.String(stmt.Where.Expr)
			queryAttributes.WhereClause = whereExpr
		}

		// Check for LIMIT clause
		if stmt.Limit != nil {
			limitValue := sqlparser.String(stmt.Limit)
			queryAttributes.Limit = limitValue
		}

	case *sqlparser.Insert:
		queryAttributes.Action = "INSERT"
		queryAttributes.TableName = sqlparser.String(stmt.Table)

		// Extract column names
		for _, col := range stmt.Columns {
			queryAttributes.Columns = append(queryAttributes.Columns, sqlparser.String(col))
		}

		// Extract values
		for _, row := range stmt.Rows.(sqlparser.Values) {
			for _, expr := range row {
				queryAttributes.Values = append(queryAttributes.Values, sqlparser.String(expr))
			}
		}

	// case *sqlparser.Update:
	// 	queryAttributes.Action = "UPDATE"
	// 	queryAttributes.TableName = sqlparser.String(stmt.Table)
	//
	// 	// Extract column names and values from SET clause
	// 	for _, assignment := range stmt.SetExprs {
	// 		queryAttributes.Values = append(queryAttributes.Values, sqlparser.String(assignment.Left))
	// 		queryAttributes.Values = append(queryAttributes.Values, sqlparser.String(assignment.Right))
	// 	}
	//
	// 	// Extract conditions from WHERE clause
	// 	if stmt.Where != nil {
	// 		queryAttributes.Conditions = sqlparser.String(stmt.Where.Expr)
	// 	}
	//
	// case *sqlparser.Delete:
	// 	queryAttributes.Action = "DELETE"
	// 	queryAttributes.TableName = sqlparser.String(stmt.Table)
	//
	// 	// Extract conditions from WHERE clause
	// 	if stmt.Where != nil {
	// 		queryAttributes.Conditions = sqlparser.String(stmt.Where.Expr)
	// 	}

	default:
		fmt.Println("Unsupported SQL statement type")
	}

	return queryAttributes
}

func main() {
	fmt.Println("Web Assembly")
	var users btree.Map[string, string]

	f := NewDatabaseService[string](&users)

	execSQL := f.SQL(&SQL{})
	js.Global().Set("execSQL", execSQL)

	// prevent the program from exiting
	select {}
}
