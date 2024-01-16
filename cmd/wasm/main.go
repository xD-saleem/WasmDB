package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/google/btree"
	"github.com/xwb1989/sqlparser"
)

// TreeNode represents a node in the btree.
type TreeNode struct {
	Key   string
	Value string
}

// BTree is a custom btree structure for holding SQL nodes.
type BTree struct {
	*btree.BTree
}

// Less implements the btree.Item interface.
func (a TreeNode) Less(b btree.Item) bool {
	return a.Key < b.(*TreeNode).Key
}

type DatabaseService struct {
	// Your DatabaseService fields go here.
}

type SQL struct {
	Query string
}

func NewDatabaseService() *DatabaseService {
	return &DatabaseService{}
}

func (ds *DatabaseService) ParseAndBuildTree(sqlQuery string) *BTree {
	tree := btree.New(2)

	r := strings.NewReader(sqlQuery)

	tokens := sqlparser.NewTokenizer(r)
	for {
		stmt, err := sqlparser.ParseNext(tokens)
		if err == io.EOF {
			break
		}
		ds.buildTree(stmt, tree)
	}

	return &BTree{tree}
}

func (ds *DatabaseService) statementType(stmt sqlparser.Statement) string {
	switch stmt.(type) {
	case *sqlparser.Select:
		return "SELECT"
	case *sqlparser.Insert:
		return "INSERT"
	case *sqlparser.Update:
		return "UPDATE"
	case *sqlparser.Delete:
		return "DELETE"
	// Add more cases for other statement types as needed.
	default:
		return "UNKNOWN"
	}
}

func (ds *DatabaseService) buildTree(stmt sqlparser.Statement, tree *btree.BTree) {
	node := &TreeNode{
		Key:   ds.statementType(stmt),
		Value: sqlparser.String(stmt),
	}

	tree.ReplaceOrInsert(node)
}

func main() {
	fmt.Println("Web Assembly")
	ds := NewDatabaseService()

	execSQL := func(query string) {
		tree := ds.ParseAndBuildTree(query)

		// Do something with the tree.
		// You can print the tree, store it, or perform operations based on your needs.
		tree.Ascend(func(item btree.Item) bool {
			node := item.(*TreeNode)
			fmt.Printf("Key: %s, Value: %s\n", node.Key, node.Value)
			return true
		})
	}

	// Example usage
	execSQL("SELECT * FROM users WHERE id = 1")

	// Prevent the program from exiting
	select {}
}
