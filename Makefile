include .env

export AUTH_SECRET
export DB_HOSTNAME
export DB_PORT
export DB_NAME
export DB_USERNAME
export DB_PASSWORD

dev:
	@air -c configs/air.toml

client:
	@go run . -c -cpuprofile=tmp/cpu.prof
	@go tool pprof -png -output tmp/out.png tmp/cpu.prof

jet:
	@jet -dsn=postgresql://${DB_USERNAME}:${DB_PASSWORD}@${DB_HOSTNAME}:${DB_PORT}/${DB_NAME}?sslmode=disable -schema=public -path=./.gen

up:
	@migrate -path ./migration -database postgresql://${DB_USERNAME}:${DB_PASSWORD}@${DB_HOSTNAME}:${DB_PORT}/${DB_NAME}?sslmode=disable up

down:
	@migrate -path ./migration -database postgresql://${DB_USERNAME}:${DB_PASSWORD}@${DB_HOSTNAME}:${DB_PORT}/${DB_NAME}?sslmode=disable down

profile:
	@go tool pprof -png -output tmp/profile.png http://localhost:8080/debug/pprof/profile

test:
	@go test -v ./core

bench:
	@go test -v -run=XXX -bench=. ./...

db:
	@docker run -p 9000:9000 -p 9009:9009 -p 8812:8812 questdb/questdb

deploy:
	@fly deploy -c configs/fly.toml
