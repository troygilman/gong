include .env

export AUTH_SECRET
export DB_HOSTNAME
export DB_PORT
export DB_NAME
export DB_USERNAME
export DB_PASSWORD

dev:
	@air

client:
	@go run . -c -cpuprofile=cpu.prof
	@go tool pprof -png -output out.png cpu.prof

migrate:
	@make up
	@make jet

jet:
	@jet -dsn=postgresql://${DB_USERNAME}:${DB_PASSWORD}@${DB_HOSTNAME}:${DB_PORT}/${DB_NAME}?sslmode=disable -schema=public -path=./.gen

up:
	@migrate -path ./migration -database postgresql://${DB_USERNAME}:${DB_PASSWORD}@${DB_HOSTNAME}:${DB_PORT}/${DB_NAME}?sslmode=disable up

down:
	@migrate -path ./migration -database postgresql://${DB_USERNAME}:${DB_PASSWORD}@${DB_HOSTNAME}:${DB_PORT}/${DB_NAME}?sslmode=disable down

profile:
	@go tool pprof -png http://localhost:8080/debug/pprof/profile > out.png

test:
	@go test -v ./domain

bench:
	@go test -v -run=XXX -bench=BenchmarkCreateEvents -o=tmp -benchmem -cpuprofile=tmp/cpu.prof -memprofile=tmp/mem.prof ./domain