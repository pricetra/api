# Docker
docker-container:
	docker compose -f ./docker-compose.yml up

# GraphQL
gql:
	go run github.com/99designs/gqlgen generate

# DB migrations
create-migration:
	go run github.com/ayaanqui/go-migration-tool --directory "./database/migrations" create-migration $(fileName)

# go-jet/jet
jet:
	go run github.com/go-jet/jet/v2/cmd/jet -dsn=postgresql://postgres:postgres@localhost:5435/postgres?sslmode=disable -path=./database/jet

test:
	go test ./tests

# Server
run:
	go run server.go
build:
	go build server.go
build-dev:
	make gql-generate && make build
watch:
	go run github.com/cosmtrek/air -c .air.toml -- -h
