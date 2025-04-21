package bind

import (
	"net/url"
	"strings"
	"sync"
)

var (
	NodeMapPool = &sync.Pool{
		New: func() any {
			return make(map[string]Node)
		},
	}
)

type Parser struct {
	nodeMapPool *sync.Pool
}

func NewParser(nodeMapPool *sync.Pool) Parser {
	return Parser{
		nodeMapPool: nodeMapPool,
	}
}

func (parser Parser) Parse(source url.Values) Node {
	node := Node{
		Children: parser.nodeMapPool.Get().(map[string]Node),
	}
	for path, val := range source {
		start := strings.Index(path, "[")
		if start == -1 {
			node.Children[path] = Node{
				Val: val[0],
			}
		} else {
			key := path[:start]
			child := node.Children[key]
			child = parser.populateNode(child, path, start, val)
			node.Children[key] = child
		}
	}
	return node

}

func (parser Parser) populateNode(node Node, path string, start int, val []string) Node {
	end := start
	for end < len(path) && path[end] != ']' {
		end++
	}

	key := path[start+1 : end]
	if node.Children == nil {
		node.Children = parser.nodeMapPool.Get().(map[string]Node)
	}

	if end == len(path)-1 {
		node.Children[key] = Node{
			Val: val[0],
		}
	} else {
		child := node.Children[key]
		child = parser.populateNode(child, path, end+1, val)
		node.Children[key] = child
	}
	return node
}
