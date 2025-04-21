package bind

import (
	"log"
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
		"people[0][first_name]": {"Troy"},
		"people[0][last_name]":  {"Gilman"},
		"people[0][email]":      {"troygilman@gmail.com"},
	}

	var data PostFormData
	assert.NoErr(t, Bind(vals, &data))

	expected := PostFormData{
		People: []Person{
			{
				FirstName: "Troy",
				LastName:  "Gilman",
				Email:     "troygilman@gmail.com",
			},
		},
	}
	assert.Equals(t, expected, data)
}

func BenchmarkBind(b *testing.B) {
	vals := url.Values{
		"people[0][first_name]": {"Troy"},
		"people[0][last_name]":  {"Gilman"},
		"people[0][email]":      {"troygilman@gmail.com"},
	}

	for range b.N {
		var data PostFormData
		if err := Bind(vals, &data); err != nil {
			b.Fatal(err)
		}
	}
}

func TestBind2(t *testing.T) {
	vals := url.Values{
		"people[0][first_name]": {"Troy"},
		"people[0][last_name]":  {"Gilman"},
		"people[0][email]":      {"troygilman@gmail.com"},
	}

	node := buildSourceNode(vals)
	log.Printf("%v", node)
}

func BenchmarkBind2(b *testing.B) {
	vals := url.Values{
		"people[0][first_name]": {"Troy"},
		"people[0][last_name]":  {"Gilman"},
		"people[0][email]":      {"troygilman@gmail.com"},
	}

	for range b.N {
		var data PostFormData
		if err := Bind2(vals, &data); err != nil {
			b.Fatal(err)
		}
	}
}
