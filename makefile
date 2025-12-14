DATABASE_URL=postgres://rwe_user:rwe_pass@localhost:5432/rwe_dev?sslmode=disable

migrate:
	migrate -path ./migrations -database "$(DATABASE_URL)" -verbose up