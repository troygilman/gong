include .env

export AUTH_SECRET

dev:
	@air

client:
	@go run . -c
