package bind

import (
	"fmt"
	"net/url"
	"slices"
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

type Node struct {
	Val      string
	Children map[string]Node
}

func (node Node) Cleanup(pool *sync.Pool) {
	if node.Children != nil {
		for key, child := range node.Children {
			child.Cleanup(pool)
			delete(node.Children, key)
		}
		pool.Put(node.Children)
	}
}

func (node Node) String() string {
	return node.stringWithIndent(0)
}

func (node Node) stringWithIndent(level int) string {
	var b strings.Builder
	indent := strings.Repeat(" ", level)

	if node.Val != "" {
		b.WriteString(node.Val)
	}

	if len(node.Children) > 0 {
		b.WriteString("\n")

		childKeys := make([]string, 0, len(node.Children))
		for key := range node.Children {
			childKeys = append(childKeys, key)
		}
		slices.Sort(childKeys)

		for _, key := range childKeys {
			child := node.Children[key]
			b.WriteString(fmt.Sprintf("%s- %s: ", indent, key))
			childStr := child.stringWithIndent(level + 1)

			// Add indentation to all lines of the child string except the first line
			lines := strings.Split(childStr, "\n")
			for j, line := range lines {
				if j == 0 {
					b.WriteString(line)
				} else {
					b.WriteString("\n" + indent + "  " + line)
				}
			}

			b.WriteString("\n")
		}
	}

	return b.String()
}

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

	// log.Printf("path: %s, start: %d, end: %d, node: %v", path, start, end, node)
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
