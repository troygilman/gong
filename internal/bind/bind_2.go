package bind

import (
	"net/url"
	"reflect"
	"strings"
)

type Node struct {
	Val      string
	Children map[string]Node
}

func Bind2(source url.Values, dest any) error {
	node := buildSourceNode(source)
	return bind(node, reflect.ValueOf(dest))
}

func buildSourceNode(source url.Values) Node {
	node := Node{
		Children: make(map[string]Node),
	}
	for key, val := range source {
		if arrayExpr.MatchString(key) {
			path := strings.Split(key, "[")
			node = populateNode(node, path, val[0])
		} else {
			node.Children[key] = Node{
				Val: val[0],
			}
		}
	}
	return node
}

func populateNode(node Node, path []string, val string) Node {
	if len(path) == 0 {
		node.Val = val
	} else {
		key := path[0]
		key = strings.TrimSuffix(key, "]")
		if node.Children == nil {
			node.Children = make(map[string]Node)
		}
		child := node.Children[key]
		child = populateNode(child, path[1:], val)
		node.Children[key] = child
	}
	return node
}

func bind(source Node, dest reflect.Value) error {
	t := dest.Type()
	switch dest.Kind() {
	case reflect.Pointer:
		return bind(source, dest.Elem())
	case reflect.Struct:
		for i := range dest.NumField() {
			field := dest.Field(i)
			if !field.CanInterface() {
				continue
			}
			tag := t.Field(i).Tag
			if sourceName, ok := tag.Lookup("form"); ok {
				_ = sourceName

			}
		}
	}
	return nil
}
