simple:
	templ generate
	go run ./example/simple

list:
	templ generate
	go run ./example/list

gen:
	templ generate

test: gen
	go test . -count=1
