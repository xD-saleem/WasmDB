package main

import (
	"fmt"
	"syscall/js"

	"github.com/tidwall/btree"
)

type Service[T interface{}] interface {
	Create() js.Func
	Get() js.Func
}

type Database[T interface{}] struct {
	tree *btree.Map[string, string]
}

func NewDatabaseService[T any](tree *btree.Map[string, string]) Service[T] {
	return Database[T]{
		tree: tree,
	}
}

func (d Database[T]) Create() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		_, isSet := d.tree.Set("user:4", "Andrea")
		if !isSet {
			return false
		}
		return true
	})
}

func (d Database[T]) Get() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		record, ok := d.tree.Get("user:4")

		var emptyValue T
		if !ok {
			return emptyValue
		}

		return record
	})
}

func main() {
	fmt.Println("Web Assembly")
	var users btree.Map[string, string]

	f := NewDatabaseService[string](&users)

	inset := f.Create()
	get := f.Get()
	js.Global().Set("insert", inset)
	js.Global().Set("get", get)
	<-make(chan bool)
}
