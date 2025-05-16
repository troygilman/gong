package bind

import (
	"net/url"
	"strings"
	"sync"
)

var (
	// NodeMapPool is a sync.Pool for recycling Node.Children maps.
	// This reduces GC pressure when parsing many form submissions.
	NodeMapPool = &sync.Pool{
		New: func() any {
			return make(map[string]Node)
		},
	}
)

// Parser parses URL form values into a tree structure for binding.
// It uses a sync.Pool to recycle maps for better performance.
type Parser struct {
	nodeMapPool *sync.Pool
}

// NewParser creates a new Parser with the provided node map pool.
// The node map pool is used to recycle maps when parsing form data.
func NewParser(nodeMapPool *sync.Pool) Parser {
	return Parser{
		nodeMapPool: nodeMapPool,
	}
}

// Parse converts URL form values into a structured Node tree.
// It handles nested form fields using bracket notation (e.g., "user[name]").
// The resulting Node tree can then be bound to Go types using Bind.
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

// populateNode recursively builds a Node tree from a form field path.
// It parses nested fields using bracket notation and assigns values at leaf nodes.
// For example, "user[name][first]" would create a nested tree of nodes.
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
