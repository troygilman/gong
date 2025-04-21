package bind

import (
	"net/url"
	"regexp"
	"strings"
	"sync"
)

var (
	ArrayExpr   = regexp.MustCompile(`^(.*?)\[([^\]]*)\](.*)$`)
	NodeMapPool = &sync.Pool{
		New: func() any {
			return make(map[string]Node)
		},
	}
)

type Node struct {
	Val         string
	Children    map[string]Node
	nodeMapPool *sync.Pool
}

func (node Node) Cleanup() {
	if node.Children != nil {
		for key, child := range node.Children {
			child.Cleanup()
			delete(node.Children, key)
		}
		node.nodeMapPool.Put(node.Children)
	}
}

type Parser struct {
	arrayExpr   *regexp.Regexp
	nodeMapPool *sync.Pool
}

func NewParser(arrayExpr *regexp.Regexp, nodeMapPool *sync.Pool) Parser {
	return Parser{
		arrayExpr:   arrayExpr,
		nodeMapPool: nodeMapPool,
	}
}

func (parser Parser) Parse(source url.Values) Node {
	node := Node{
		Children:    parser.nodeMapPool.Get().(map[string]Node),
		nodeMapPool: parser.nodeMapPool,
	}
	for key, val := range source {
		if parser.arrayExpr.MatchString(key) {
			path := strings.Split(key, "[")
			node = parser.populateNode(node, path, val[0])
		} else {
			node.Children[key] = Node{
				Val: val[0],
			}
		}
	}
	return node

}

func (parser Parser) populateNode(node Node, path []string, val string) Node {
	if len(path) == 0 {
		node.Val = val
	} else {
		key := path[0]
		key = strings.TrimSuffix(key, "]")
		if node.Children == nil {
			node.Children = parser.nodeMapPool.Get().(map[string]Node)
			node.nodeMapPool = parser.nodeMapPool
		}
		child := node.Children[key]
		child = parser.populateNode(child, path[1:], val)
		node.Children[key] = child
	}
	return node
}
