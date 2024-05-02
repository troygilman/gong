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
	@go run . -c

migrate:
	@make up
	@make jet

jet:
	@jet -dsn=postgresql://${DB_USERNAME}:${DB_PASSWORD}@${DB_HOSTNAME}:${DB_PORT}/${DB_NAME}?sslmode=disable -schema=public -path=./.gen

up:
	@migrate -path ./migration -database postgresql://${DB_USERNAME}:${DB_PASSWORD}@${DB_HOSTNAME}:${DB_PORT}/${DB_NAME}?sslmode=disable up

down:
	@migrate -path ./migration -database postgresql://${DB_USERNAME}:${DB_PASSWORD}@${DB_HOSTNAME}:${DB_PORT}/${DB_NAME}?sslmode=disable down
