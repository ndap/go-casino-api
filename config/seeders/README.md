# Database Seeders

Folder ini berisi seeder untuk mengisi database dengan data awal yang diperlukan untuk testing dan development.

## File Seeder

### 1. `userSeeders.go`

- Membuat user admin dan test user
- Membuat wallet untuk setiap user
- Default credentials:
  - Admin: `admin@casino.com` / `admin123`
  - User: `user@casino.com` / `user123`

### 2. `gameSeeders.go`

- Membuat game settings default
- Konfigurasi:
  - Max Multiplier: 100.0x
  - Min Bet: 1,000 IDR
  - Max Bet: 1,000,000 IDR
  - Multiplier Speed: 0.1 per detik

### 3. `seeders.go`

- File utama yang menjalankan semua seeder
- Fungsi `SeedAllData()` untuk menjalankan semua seeder

## Cara Menjalankan Seeder

### 1. Menjalankan Semua Seeder

```bash
make seed
# atau
go run cmd/seeder/main.go
```

### 2. Menjalankan Seeder Tertentu (dari kode)

```go
import "casino_api_go/config/seeders"

// Seed users saja
seeders.SeedUsers()

// Seed game settings saja
seeders.SeedAllGameData()

// Seed semua data
seeders.SeedAllData()
```

## Data yang Dibuat

### Users

- **Admin User**

  - Username: `admin`
  - Email: `admin@casino.com`
  - Password: `admin123`
  - Role: `admin`
  - Wallet: 1,000,000 IDR

- **Test User**
  - Username: `testuser`
  - Email: `user@casino.com`
  - Password: `user123`
  - Role: `user`
  - Wallet: 100,000 IDR

### Game Settings

- Max Multiplier: 100.0x
- Min Bet Amount: 1,000 IDR
- Max Bet Amount: 1,000,000 IDR
- Multiplier Speed: 0.1 per detik
- Status: Active

## Catatan

- Seeder akan mengecek apakah data sudah ada sebelum membuat data baru
- Jika data sudah ada, seeder akan di-skip
- Pastikan database sudah terhubung sebelum menjalankan seeder
