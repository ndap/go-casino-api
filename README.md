# Casino API Go

API backend untuk aplikasi casino online yang dibangun dengan Go dan Gin framework.

## 🚀 Fitur Utama

- **Autentikasi & Otorisasi**: Login, register, JWT token, role-based access
- **Sistem Wallet**: Deposit, withdraw, top-up, dan riwayat transaksi
- **Game Casino**: Crash game dengan sistem multiplier dan betting
- **Admin Panel**: Manajemen user, game settings, dan dashboard
- **Database**: MySQL dengan GORM ORM

## 📋 Persyaratan

- Go 1.24.5 atau lebih baru
- MySQL 5.7 atau lebih baru
- Git

## 🛠️ Instalasi

1. **Clone repository**

   ```bash
   git clone <repository-url>
   cd casino_api_go
   ```

2. **Install dependencies**

   ```bash
   go mod tidy
   ```

3. **Setup database**

   - Buat database MySQL dengan nama `casino_api_go`
   - Copy file `env.example` menjadi `.env`
   - Edit konfigurasi database di file `.env`

4. **Jalankan aplikasi**

   ```bash
   go run main.go
   ```

   Atau menggunakan Makefile:

   ```bash
   make run
   ```

## 🔧 Konfigurasi

File `.env`:

```env
DB_USER=root
DB_PASS=
DB_HOST=localhost
DB_PORT=3306
DB_NAME=casino_api_go
JWT_SECRET=your-super-secret-jwt-key-change-in-production
```

## 📚 API Endpoints

### Autentikasi

- `POST /api/auth/register` - Register user baru
- `POST /api/auth/login` - Login user

### User (Protected)

- `GET /api/profile` - Get profile user
- `PUT /api/profile` - Update profile
- `PUT /api/change-password` - Ganti password
- `POST /api/logout` - Logout
- `GET /api/wallet` - Get wallet info
- `POST /api/deposit` - Deposit saldo
- `POST /api/withdraw` - Withdraw saldo
- `GET /api/transactions` - Riwayat transaksi

### Casino Game (Protected)

- `POST /api/casino/start` - Mulai game
- `POST /api/casino/stop` - Stop game
- `GET /api/casino/games` - Daftar game user
- `GET /api/casino/settings` - Game settings
- `GET /api/casino/active-games` - Status game aktif
- `GET /api/casino/game/:id` - Status game tertentu

### Admin (Admin Only)

- `GET /api/admin/dashboard` - Dashboard admin
- `GET /api/admin/users` - Daftar semua user
- `POST /api/admin/users/:id/ban` - Ban user
- `POST /api/admin/users/:id/unban` - Unban user
- `POST /api/admin/users/:id/wallet/topup` - Top-up wallet user
- `GET /api/admin/games` - Daftar semua game
- `PUT /api/admin/game-settings` - Update game settings

## 🗄️ Database Schema

### Models

- **User**: Username, email, password, role, status
- **Wallet**: Balance, currency, user_id
- **Game**: Bet amount, multiplier, win amount, crash point, status
- **Transaction**: Type, amount, balance, description, status
- **GameSettings**: Max multiplier, min/max bet, speed settings

## 🎮 Game Mechanics

- **Crash Game**: Game dengan sistem multiplier yang naik secara eksponensial
- **Betting**: User dapat bet sebelum game dimulai
- **Cash Out**: User dapat cash out kapan saja sebelum crash
- **Win/Loss**: Jika user cash out sebelum crash = win, jika tidak = loss

## 🛠️ Development

### Available Commands

```bash
make help          # Lihat semua command yang tersedia
make run           # Jalankan aplikasi
make build         # Build aplikasi
make test          # Jalankan test
make seed          # Jalankan seeder
make clean         # Bersihkan build files
```

### Database Seeding

Aplikasi akan otomatis menjalankan seeder saat startup untuk:

- Membuat user admin default
- Setup game settings default
- Membuat sample data

## 🔒 Security Features

- JWT token authentication
- Password hashing dengan bcrypt
- CORS enabled
- Role-based access control
- Token blacklisting untuk logout

## 📝 Health Check

- `GET /health` - Check status API

## 🚀 Production

1. Update JWT_SECRET di file `.env`
2. Setup database production
3. Build aplikasi: `make build`
4. Deploy binary ke server

## 📄 License

MIT License
