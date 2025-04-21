package bind

import (
	"net/url"
	"testing"

	"github.com/troygilman/gong/internal/assert"
)

type PostFormData struct {
	People []Person `form:"people"`
}

type Person struct {
	FirstName string `form:"first_name"`
	LastName  string `form:"last_name"`
	Email     string `form:"email"`
}

func TestBind(t *testing.T) {
	vals := url.Values{
		"people[0][first_name]": {"Bob"},
		"people[0][last_name]":  {"Ross"},
		"people[0][email]":      {"bobross@gmail.com"},
	}

	var data PostFormData
	assert.NoErr(t, Bind(vals, &data))

	expected := PostFormData{
		People: []Person{
			{
				FirstName: "Bob",
				LastName:  "Ross",
				Email:     "bobross@gmail.com",
			},
		},
	}
	assert.Equals(t, expected, data)
}

func TestBindMap(t *testing.T) {
	vals := url.Values{
		"people[0][first_name]": {"Bob"},
		"people[0][last_name]":  {"Ross"},
		"people[0][email]":      {"bobross@gmail.com"},
	}

	var data map[string]any
	assert.NoErr(t, Bind(vals, &data))

	expected := map[string]any{
		"people": map[string]any{
			"0": map[string]any{
				"first_name": "Bob",
				"last_name":  "Ross",
				"email":      "bobross@gmail.com",
			},
		},
	}
	assert.Equals(t, expected, data)
}

func BenchmarkBind(b *testing.B) {
	vals := url.Values{
		"people[0][first_name]": {"Bob"},
		"people[0][last_name]":  {"Ross"},
		"people[0][email]":      {"bobross@gmail.com"},
	}

	for b.Loop() {
		var data PostFormData
		if err := Bind(vals, &data); err != nil {
			b.Fatal(err)
		}
	}
}
