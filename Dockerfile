ARG GO_VERSION=1
ARG NODE_VERSION=18
FROM golang:${GO_VERSION}-bookworm as builder

# Set up the Go environment
WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
RUN go install github.com/a-h/templ/cmd/templ@latest

# Stage to handle Node.js dependencies and build CSS using Tailwind
FROM node:${NODE_VERSION}-bookworm-slim as node-builder

WORKDIR /usr/src/app
COPY . .
RUN npm install
RUN npx tailwindcss build -c configs/tailwind.config.js -i configs/tailwind.css -o static/css/app.css

# Bring the Node.js built assets into the Go builder stage
FROM builder as final-builder

WORKDIR /usr/src/app
COPY --from=node-builder /usr/src/app/static /usr/src/app/static
COPY . .

RUN templ generate
RUN go build -v -o /run-app .

# Final runtime image
FROM debian:bookworm

COPY --from=final-builder /run-app /usr/local/bin/
CMD ["run-app"]
