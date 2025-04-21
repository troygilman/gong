package bind

import (
	"net/url"
	"testing"

	"github.com/troygilman/gong/internal/assert"
)

func TestParser_Parse(t *testing.T) {
	// Create a parser instance
	parser := NewParser(ArrayExpr, NodeMapPool)

	// Test case: nested form data like user[name]=John&user[address][city]=New York
	testValues := url.Values{
		"user[name]":           []string{"John"},
		"user[address][city]":  []string{"New York"},
		"user[address][state]": []string{"NY"},
		"simple":               []string{"value"},
	}

	// Parse the values
	result := parser.Parse(testValues)
	defer result.Cleanup(NodeMapPool)

	expected := Node{
		Val: "",
		Children: map[string]Node{
			"user": {
				Val: "",
				Children: map[string]Node{
					"name": {
						Val: "John",
					},
					"address": {
						Val: "",
						Children: map[string]Node{
							"city": {
								Val: "New York",
							},
							"state": {
								Val: "NY",
							},
						},
					},
				},
			},
			"simple": {
				Val: "value",
			},
		},
	}

	assert.Equals(t, expected, result)
}
