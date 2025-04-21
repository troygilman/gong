package bind

import (
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/troygilman/gong/internal/assert"
)

type PostFormData struct {
	People []Person `form:"people"`
}

type Person struct {
	FirstName string  `form:"first_name"`
	LastName  string  `form:"last_name"`
	Email     string  `form:"email"`
	Age       int     `form:"age"`
	Active    bool    `form:"active"`
	Score     float64 `form:"score"`
}

type NestedStruct struct {
	Person Person `form:"person"`
	Role   string `form:"role"`
}

type ComplexStruct struct {
	People    []Person           `form:"people"`
	Settings  map[string]string  `form:"settings"`
	Metadata  map[string]any     `form:"metadata"`
	CreatedAt time.Time          `form:"created_at"`
	Tags      []string           `form:"tags"`
	Counts    map[string]int     `form:"counts"`
	Ratings   map[string]float64 `form:"ratings"`
	Flags     map[string]bool    `form:"flags"`
	IntKeyMap map[int]string     `form:"int_key_map"`
	Nested    NestedStruct       `form:"nested"`
}

type BasicTypes struct {
	Name      string    `form:"name"`
	Age       int       `form:"age"`
	Score     float64   `form:"score"`
	Active    bool      `form:"active"`
	CreatedAt time.Time `form:"created_at"`
}

func TestBindValidation(t *testing.T) {
	tests := []struct {
		name    string
		values  url.Values
		dest    any
		wantErr bool
	}{
		{
			name:    "non-pointer struct",
			values:  url.Values{},
			dest:    Person{},
			wantErr: false,
		},
		{
			name:    "non-pointer map",
			values:  url.Values{},
			dest:    map[string]string{},
			wantErr: false,
		},
		{
			name:    "pointer to non-struct/non-map",
			values:  url.Values{},
			dest:    new(string),
			wantErr: false,
		},
		{
			name:    "nil pointer",
			values:  url.Values{},
			dest:    (*Person)(nil),
			wantErr: true,
		},
		{
			name:    "valid struct pointer",
			values:  url.Values{},
			dest:    new(Person),
			wantErr: false,
		},
		{
			name:    "valid map pointer",
			values:  url.Values{},
			dest:    new(map[string]string),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Bind(tt.values, tt.dest)
			if tt.wantErr {
				assert.Err(t, err)
			} else {
				assert.NoErr(t, err)
			}
		})
	}
}

func TestBindBasicTypes(t *testing.T) {
	tests := []struct {
		name    string
		values  url.Values
		dest    any
		want    any
		wantErr bool
	}{
		{
			name: "string value",
			values: url.Values{
				"name": {"John"},
			},
			dest: new(BasicTypes),
			want: BasicTypes{
				Name: "John",
			},
		},
		{
			name: "int value",
			values: url.Values{
				"age": {"30"},
			},
			dest: new(BasicTypes),
			want: BasicTypes{
				Age: 30,
			},
		},
		{
			name: "float value",
			values: url.Values{
				"score": {"95.5"},
			},
			dest: new(BasicTypes),
			want: BasicTypes{
				Score: 95.5,
			},
		},
		{
			name: "bool value",
			values: url.Values{
				"active": {"true"},
			},
			dest: new(BasicTypes),
			want: BasicTypes{
				Active: true,
			},
		},
		{
			name: "time value",
			values: url.Values{
				"created_at": {"2024-03-20T15:04:05Z"},
			},
			dest: new(BasicTypes),
			want: BasicTypes{
				CreatedAt: time.Date(2024, 3, 20, 15, 4, 5, 0, time.UTC),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Bind(tt.values, tt.dest)
			if tt.wantErr {
				assert.Err(t, err)
				return
			}
			assert.NoErr(t, err)
			assert.Equals(t, tt.want, reflect.ValueOf(tt.dest).Elem().Interface())
		})
	}
}

func TestBindStruct(t *testing.T) {
	tests := []struct {
		name    string
		values  url.Values
		dest    any
		want    any
		wantErr bool
	}{
		{
			name: "basic struct",
			values: url.Values{
				"first_name": {"John"},
				"last_name":  {"Doe"},
				"email":      {"john@example.com"},
				"age":        {"30"},
				"active":     {"true"},
				"score":      {"95.5"},
			},
			dest: new(Person),
			want: Person{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john@example.com",
				Age:       30,
				Active:    true,
				Score:     95.5,
			},
		},
		{
			name: "nested struct",
			values: url.Values{
				"person[first_name]": {"John"},
				"person[last_name]":  {"Doe"},
				"person[email]":      {"john@example.com"},
				"role":               {"admin"},
			},
			dest: new(NestedStruct),
			want: NestedStruct{
				Person: Person{
					FirstName: "John",
					LastName:  "Doe",
					Email:     "john@example.com",
				},
				Role: "admin",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Bind(tt.values, tt.dest)
			if tt.wantErr {
				assert.Err(t, err)
				return
			}
			assert.NoErr(t, err)
			assert.Equals(t, tt.want, reflect.ValueOf(tt.dest).Elem().Interface())
		})
	}
}

func TestBindMap(t *testing.T) {
	tests := []struct {
		name    string
		values  url.Values
		dest    any
		want    any
		wantErr bool
	}{
		{
			name: "string map",
			values: url.Values{
				"settings[theme]":  {"dark"},
				"settings[locale]": {"en-US"},
			},
			dest: new(map[string]any),
			want: map[string]any{
				"settings": map[string]any{
					"theme":  "dark",
					"locale": "en-US",
				},
			},
		},
		{
			name: "int map",
			values: url.Values{
				"counts[users]":    {"100"},
				"counts[posts]":    {"500"},
				"counts[comments]": {"1000"},
			},
			dest: new(map[string]map[string]int),
			want: map[string]map[string]int{
				"counts": {
					"users":    100,
					"posts":    500,
					"comments": 1000,
				},
			},
		},
		{
			name: "float map",
			values: url.Values{
				"ratings[quality]": {"4.5"},
				"ratings[speed]":   {"3.8"},
			},
			dest: new(map[string]map[string]float64),
			want: map[string]map[string]float64{
				"ratings": {
					"quality": 4.5,
					"speed":   3.8,
				},
			},
		},
		{
			name: "bool map",
			values: url.Values{
				"flags[enabled]":  {"true"},
				"flags[verified]": {"false"},
			},
			dest: new(map[string]map[string]bool),
			want: map[string]map[string]bool{
				"flags": {
					"enabled":  true,
					"verified": false,
				},
			},
		},
		{
			name: "int key map",
			values: url.Values{
				"int_key_map[1]": {"one"},
				"int_key_map[2]": {"two"},
			},
			dest: new(map[string]map[int]string),
			want: map[string]map[int]string{
				"int_key_map": {
					1: "one",
					2: "two",
				},
			},
		},
		{
			name: "interface map",
			values: url.Values{
				"metadata[name]":          {"John"},
				"metadata[age]":           {"30"},
				"metadata[active]":        {"true"},
				"metadata[score]":         {"95.5"},
				"metadata[nested][value]": {"deep"},
			},
			dest: new(map[string]any),
			want: map[string]any{
				"metadata": map[string]any{
					"name":   "John",
					"age":    "30",
					"active": "true",
					"score":  "95.5",
					"nested": map[string]any{
						"value": "deep",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Bind(tt.values, tt.dest)
			if tt.wantErr {
				assert.Err(t, err)
				return
			}
			assert.NoErr(t, err)
			assert.Equals(t, tt.want, reflect.ValueOf(tt.dest).Elem().Interface())
		})
	}
}

func TestBindComplex(t *testing.T) {
	vals := url.Values{
		"people[0][first_name]":      {"John"},
		"people[0][last_name]":       {"Doe"},
		"people[0][email]":           {"john@example.com"},
		"settings[theme]":            {"dark"},
		"metadata[name]":             {"John"},
		"metadata[age]":              {"30"},
		"created_at":                 {"2024-03-20T15:04:05Z"},
		"tags[0]":                    {"go"},
		"tags[1]":                    {"testing"},
		"counts[users]":              {"100"},
		"ratings[quality]":           {"4.5"},
		"flags[enabled]":             {"true"},
		"int_key_map[1]":             {"one"},
		"nested[person][first_name]": {"Jane"},
		"nested[role]":               {"admin"},
	}

	var data ComplexStruct
	assert.NoErr(t, Bind(vals, &data))

	expected := ComplexStruct{
		People: []Person{
			{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john@example.com",
			},
		},
		Settings: map[string]string{
			"theme": "dark",
		},
		Metadata: map[string]any{
			"name": "John",
			"age":  "30",
		},
		CreatedAt: time.Date(2024, 3, 20, 15, 4, 5, 0, time.UTC),
		Tags:      []string{"go", "testing"},
		Counts: map[string]int{
			"users": 100,
		},
		Ratings: map[string]float64{
			"quality": 4.5,
		},
		Flags: map[string]bool{
			"enabled": true,
		},
		IntKeyMap: map[int]string{
			1: "one",
		},
		Nested: NestedStruct{
			Person: Person{
				FirstName: "Jane",
			},
			Role: "admin",
		},
	}

	assert.Equals(t, expected, data)
}

func TestBindErrors(t *testing.T) {
	tests := []struct {
		name    string
		values  url.Values
		dest    any
		wantErr bool
	}{
		{
			name: "invalid int",
			values: url.Values{
				"age": {"notanint"},
			},
			dest:    new(Person),
			wantErr: true,
		},
		{
			name: "invalid float",
			values: url.Values{
				"score": {"notafloat"},
			},
			dest:    new(Person),
			wantErr: true,
		},
		{
			name: "invalid bool",
			values: url.Values{
				"active": {"notabool"},
			},
			dest:    new(Person),
			wantErr: true,
		},
		{
			name: "invalid time",
			values: url.Values{
				"created_at": {"notatime"},
			},
			dest:    new(ComplexStruct),
			wantErr: true,
		},
		{
			name: "invalid slice index",
			values: url.Values{
				"tags[notanindex]": {"value"},
			},
			dest:    new(ComplexStruct),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Bind(tt.values, tt.dest)
			if tt.wantErr {
				assert.Err(t, err)
			} else {
				assert.NoErr(t, err)
			}
		})
	}
}

func BenchmarkBind(b *testing.B) {
	vals := url.Values{
		"people[0][first_name]": {"Bob"},
		"people[0][last_name]":  {"Ross"},
		"people[0][email]":      {"bobross@gmail.com"},
	}

	for b.Loop() {
		var data ComplexStruct
		if err := Bind(vals, &data); err != nil {
			b.Fatal(err)
		}
	}
}
