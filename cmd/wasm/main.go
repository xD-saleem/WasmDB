package main

import (
	"fmt"
	"syscall/js"

	"github.com/google/btree"
	"github.com/xwb1989/sqlparser"
)

type Action struct {
	Type         string   // Type of action (e.g., "SELECT", "INSERT", "UPDATE", "DELETE")
	TableNames   []string // Tables involved in the action
	ColumnNames  []string // Columns involved in the action (if applicable)
	Values       []string // Values used in the action (if applicable)
	Conditions   []string // Conditions specified in the action (if applicable)
	DDLStatement string   // Full DDL statement (if applicable)
}

// BTree is a custom btree structure for holding SQL nodes.

type DatabaseService struct {
	Btree *btree.BTree
}

type SQL struct {
	Query string
}

func NewDatabaseService(t *btree.BTree) *DatabaseService {
	return &DatabaseService{
		Btree: t,
	}
}

type Node struct {
	val dynamicValue
	Id  int
}

// Less is required to implement the btree.Item interface.
func (a Node) Less(b btree.Item) bool {
	return a.Id < b.(Node).Id
}

func (ds *DatabaseService) execSql(sql string) {
	stmt, err := sqlparser.Parse(sql)
	if err != nil {
		fmt.Println("Error parsing SQL:", err)
	}

	action := ds.buildAction(stmt)
	if action != nil {
		ds.execAction(action)
	}
}

func main() {
	fmt.Println("Web Assembly")
	treePtr := btree.New(2)
	ds := NewDatabaseService(treePtr)

	execSQL := js.FuncOf(func(this js.Value, p []js.Value) interface{} {
		// Ensure that the function is called with at least one argument
		if len(p) < 1 {
			return nil
		}

		// Get the SQL query from the first argument
		query := p[0].String()

		// Call the execSql method with the query
		ds.execSql(query)

		// Return nil or an appropriate value based on your requirements
		return nil
	})

	// Expose execSQL to JavaScript
	js.Global().Set("execSQL", execSQL)

	// Example usage
	// execSQL("INSERT INTO Customers (CustomerName, ContactName, Address, City, PostalCode, Country) VALUES ('Cardinal', 'Tom B. Erichsen', 'Skagen 21', 'Stavanger', '4006', 'Norway');")

	// execSQL("SELECT * FROM Customer")

	// Prevent the program from exiting
	select {}
}
