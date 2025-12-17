DATABASE_URL=postgres://rwe_user:rwe_pass@localhost:5432/rwe_dev?sslmode=disable

create-migration:
	migrate create -ext sql -dir ./migrations -seq $(MIGRATION_NAME)

migrate:
	migrate -path ./migrations -database "$(DATABASE_URL)" -verbose up