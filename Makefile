# Makefile untuk Casino API Go

.PHONY: help run build test seed seed-users seed-games clean

# Default target
help:
	@echo "Available commands:"
	@echo "  run        - Menjalankan aplikasi"
	@echo "  build      - Build aplikasi"
	@echo "  test       - Menjalankan test"
	@echo "  seed       - Menjalankan semua seeder"
	@echo "  seed-users - Menjalankan seeder user saja"
	@echo "  seed-games - Menjalankan seeder game saja"
	@echo "  clean      - Membersihkan build files"

# Menjalankan aplikasi
run:
	go run main.go

# Build aplikasi
build:
	go build -o bin/casino-api main.go

# Menjalankan test
test:
	go test ./...

# Menjalankan semua seeder
seed:
	go run cmd/seeder/main.go

# Menjalankan seeder user saja (dari kode)
seed-users:
	@echo "Untuk menjalankan seeder user saja, tambahkan kode berikut di main.go:"
	@echo "import 'casino_api_go/config/seeders'"
	@echo "seeders.SeedUsers()"

# Menjalankan seeder game saja (dari kode)
seed-games:
	@echo "Untuk menjalankan seeder game saja, tambahkan kode berikut di main.go:"
	@echo "import 'casino_api_go/config/seeders'"
	@echo "seeders.SeedAllGameData()"

# Membersihkan build files
clean:
	rm -rf bin/
	go clean

# Install dependencies
deps:
	go mod tidy
	go mod download
