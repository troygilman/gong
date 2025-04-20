package bind

import "net/http"

func Bind(request *http.Request, dest any) error {
	if err := request.ParseForm(); err != nil {
		return err
	}
	return nil
}
