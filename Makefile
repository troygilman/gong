simple:
	templ generate
	go run ./example/simple

list:
	templ generate
	go run ./example/list

build:
	templ generate

test: build
	go test . ./**/* -count=1 -v
