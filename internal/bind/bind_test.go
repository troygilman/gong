package bind

import (
	"net/http"
	"strings"
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
	reader := strings.NewReader("people%5B1%5D%5Bfirst_name%5D=Troy&people%5B1%5D%5Blast_name%5D=Gilman&people%5B1%5D%5Bemail%5D=troygilman%40gmail.com")
	r, err := http.NewRequest(http.MethodGet, "/", reader)
	assert.NoErr(t, err)

	var data PostFormData
	err = Bind(r, &data)
	assert.NoErr(t, err)

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
