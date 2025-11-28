DB_URL=postgres://postgres:postgres@localhost:5432/auth_db?sslmode=disable

run:
	go run cmd/server/main.go