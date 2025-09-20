# Casino API Go

API backend untuk aplikasi casino yang dibangun dengan Go, Gin, dan GORM.

## Fitur

- ✅ Authentication dengan JWT
- ✅ User registration dan login
- ✅ Password hashing dengan bcrypt
- ✅ User profile management
- ✅ Wallet system
- ✅ Database migration otomatis
- ✅ CORS support
- ✅ Input validation
- ✅ Error handling yang konsisten

## Setup

### 1. Prerequisites

- Go 1.24+
- MySQL/MariaDB
- Laragon (untuk development)

### 2. Environment Variables

Buat file `.env` di root directory:

```env
DB_USER=root
DB_PASS=
DB_HOST=localhost
DB_PORT=3306
DB_NAME=casino_api_go
JWT_SECRET=your-super-secret-jwt-key-change-in-production
```

### 3. Install Dependencies

```bash
go mod tidy
```

### 4. Run Application

```bash
go run main.go
```

Server akan berjalan di `http://localhost:8080`

## API Endpoints

### Authentication

#### Register User

```http
POST /api/auth/register
Content-Type: application/json

{
  "username": "john_doe",
  "email": "john@example.com",
  "password": "password123"
}
```

**Response:**

```json
{
  "success": true,
  "message": "User registered successfully",
  "data": {
    "user": {
      "id": 1,
      "username": "john_doe",
      "email": "john@example.com",
      "role": "user",
      "status": "active"
    },
    "wallet": {
      "balance": 0,
      "currency": "IDR"
    }
  },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

#### Login User

```http
POST /api/auth/login
Content-Type: application/json

{
  "email": "john@example.com",
  "password": "password123"
}
```

**Response:**

```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "user": {
      "id": 1,
      "username": "john_doe",
      "email": "john@example.com",
      "role": "user",
      "status": "active"
    },
    "wallet": {
      "balance": 0,
      "currency": "IDR"
    }
  },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### Protected Endpoints

Semua endpoint di bawah ini memerlukan header Authorization:

```
Authorization: Bearer <your-jwt-token>
```

#### Get Profile

```http
GET /api/profile
Authorization: Bearer <token>
```

#### Update Profile

```http
PUT /api/profile
Authorization: Bearer <token>
Content-Type: application/json

{
  "username": "new_username",
  "email": "newemail@example.com"
}
```

#### Change Password

```http
PUT /api/change-password
Authorization: Bearer <token>
Content-Type: application/json

{
  "current_password": "oldpassword",
  "new_password": "newpassword123"
}
```

#### Get Wallet

```http
GET /api/wallet
Authorization: Bearer <token>
```

## 🎮 **Casino Game Endpoints**

### Game Management

- `POST /api/casino/start` - Start new game
- `POST /api/casino/stop` - Stop active game
- `GET /api/casino/game/:id` - Get game status
- `GET /api/casino/games` - Get user games
- `GET /api/casino/settings` - Get game settings

### Crash Monitoring (NEW)

- `GET /api/casino/active-games` - Get status of all active games
- `GET /api/casino/game/:id/crash-info` - Get crash information for specific game

### Authentication

Semua endpoint casino memerlukan authentication dengan Bearer token.

### Admin Endpoints

Semua endpoint admin memerlukan role admin dan header Authorization:

```
Authorization: Bearer <admin-jwt-token>
```

#### Get Dashboard Statistics

```http
GET /api/admin/dashboard
Authorization: Bearer <admin-token>
```

#### Get All Users (with pagination)

```http
GET /api/admin/users?page=1&limit=10&search=john&status=active
Authorization: Bearer <admin-token>
```

#### Get User by ID

```http
GET /api/admin/users/1
Authorization: Bearer <admin-token>
```

#### Ban User

```http
POST /api/admin/users/1/ban
Authorization: Bearer <admin-token>
Content-Type: application/json

{
  "reason": "Violation of terms of service"
}
```

#### Unban User

```http
POST /api/admin/users/1/unban
Authorization: Bearer <admin-token>
```

#### Top Up User Wallet

```http
POST /api/admin/users/1/wallet/topup
Authorization: Bearer <admin-token>
Content-Type: application/json

{
  "amount": 1000000,
  "currency": "IDR",
  "note": "Bonus for new user"
}
```

#### Deduct from User Wallet

```http
POST /api/admin/users/1/wallet/deduct
Authorization: Bearer <admin-token>
Content-Type: application/json

{
  "amount": 50000,
  "currency": "IDR",
  "note": "Service fee"
}
```

#### Get User Wallet History

```http
GET /api/admin/users/1/wallet/history
Authorization: Bearer <admin-token>
```

#### Get All Games

```http
GET /api/admin/games?page=1&limit=10&status=won
Authorization: Bearer <admin-token>
```

#### Get Game Settings

```http
GET /api/admin/game-settings
Authorization: Bearer <admin-token>
```

#### Update Game Settings

```http
PUT /api/admin/game-settings
Authorization: Bearer <admin-token>
Content-Type: application/json

{
  "max_multiplier": 100.0,
  "win_percentage": 30.0,
  "min_bet_amount": 1000,
  "max_bet_amount": 1000000,
  "multiplier_speed": 0.1,
  "house_edge": 5.0,
  "is_active": true
}
```

### Health Check

```http
GET /health
```

## Database Schema

### Users Table

```sql
CREATE TABLE users (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  username VARCHAR(255) NOT NULL UNIQUE,
  password VARCHAR(255) NOT NULL,
  email VARCHAR(255) NOT NULL UNIQUE,
  role ENUM('admin', 'user') DEFAULT 'user',
  status ENUM('active', 'banned') DEFAULT 'active',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

### Wallets Table

```sql
CREATE TABLE wallets (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id BIGINT UNSIGNED NOT NULL UNIQUE,
  balance DECIMAL(15,2) NOT NULL DEFAULT 0,
  currency VARCHAR(10) NOT NULL DEFAULT 'IDR',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

### Games Table

```sql
CREATE TABLE games (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id BIGINT UNSIGNED NOT NULL,
  bet_amount DECIMAL(15,2) NOT NULL,
  multiplier DECIMAL(10,2) NOT NULL DEFAULT 1.0,
  win_amount DECIMAL(15,2) NOT NULL DEFAULT 0,
  status ENUM('active', 'won', 'lost') DEFAULT 'active',
  is_completed BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

### Game Settings Table

```sql
CREATE TABLE game_settings (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  max_multiplier DECIMAL(10,2) NOT NULL DEFAULT 100.0,
  win_percentage DECIMAL(5,2) NOT NULL DEFAULT 30.0,
  min_bet_amount DECIMAL(15,2) NOT NULL DEFAULT 1000,
  max_bet_amount DECIMAL(15,2) NOT NULL DEFAULT 1000000,
  multiplier_speed DECIMAL(5,2) NOT NULL DEFAULT 0.1,
  house_edge DECIMAL(5,2) NOT NULL DEFAULT 5.0,
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

### Transactions Table

```sql
CREATE TABLE transactions (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id BIGINT UNSIGNED NOT NULL,
  game_id BIGINT UNSIGNED NULL,
  type ENUM('bet', 'win', 'loss', 'topup', 'deduct') NOT NULL,
  amount DECIMAL(15,2) NOT NULL,
  balance DECIMAL(15,2) NOT NULL,
  description VARCHAR(255) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  FOREIGN KEY (game_id) REFERENCES games(id) ON DELETE SET NULL
);
```

## Error Responses

Semua error response mengikuti format yang konsisten:

```json
{
  "success": false,
  "message": "Error description"
}
```

### Common HTTP Status Codes

- `200` - Success
- `201` - Created
- `400` - Bad Request (validation error)
- `401` - Unauthorized (invalid token)
- `403` - Forbidden (account banned)
- `404` - Not Found
- `409` - Conflict (duplicate data)
- `500` - Internal Server Error

## Security Features

- Password hashing dengan bcrypt
- JWT token authentication
- Input validation
- SQL injection protection (GORM)
- CORS configuration
- Environment variable untuk sensitive data

## Development

### Project Structure

```
casino_api_go/
├── config/
│   └── database.go
├── controllers/
│   ├── auth.go
│   ├── user.go
│   ├── admin.go
│   └── casino.go
├── models/
│   ├── user.go
│   ├── wallet.go
│   ├── game.go
│   └── transaction.go
├── routes/
│   ├── auth.go
│   ├── admin.go
│   └── casino.go
├── main.go
├── go.mod
├── go.sum
└── .env
```

### Testing API

Gunakan tools seperti Postman atau curl untuk testing:

```bash
# Register
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","email":"test@example.com","password":"password123"}'

# Login
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'

# Get Profile (with token)
curl -X GET http://localhost:8080/api/profile \
  -H "Authorization: Bearer <your-token>"

# Admin endpoints (require admin token)
curl -X GET http://localhost:8080/api/admin/dashboard \
  -H "Authorization: Bearer <admin-token>"

curl -X GET http://localhost:8080/api/admin/users \
  -H "Authorization: Bearer <admin-token>"

curl -X POST http://localhost:8080/api/admin/users/1/ban \
  -H "Authorization: Bearer <admin-token>" \
  -H "Content-Type: application/json" \
  -d '{"reason":"Violation of terms"}'

curl -X POST http://localhost:8080/api/admin/users/1/wallet/topup \
  -H "Authorization: Bearer <admin-token>" \
  -H "Content-Type: application/json" \
  -d '{"amount":1000000,"currency":"IDR","note":"Bonus"}'

# Casino game endpoints
curl -X POST http://localhost:8080/api/casino/start \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"bet_amount":10000}'

curl -X POST http://localhost:8080/api/casino/stop \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"game_id":1}'

curl -X GET http://localhost:8080/api/casino/game/1 \
  -H "Authorization: Bearer <token>"

# Admin casino endpoints
curl -X GET http://localhost:8080/api/admin/games \
  -H "Authorization: Bearer <admin-token>"

curl -X PUT http://localhost:8080/api/admin/game-settings \
  -H "Authorization: Bearer <admin-token>" \
  -H "Content-Type: application/json" \
  -d '{"max_multiplier":100.0,"win_percentage":30.0,"min_bet_amount":1000,"max_bet_amount":1000000,"multiplier_speed":0.1,"house_edge":5.0,"is_active":true}'
```

## Production Deployment

1. Ganti JWT secret dengan key yang kuat
2. Gunakan environment variables untuk semua sensitive data
3. Setup database dengan user yang memiliki permission terbatas
4. Enable HTTPS
5. Setup logging dan monitoring
6. Implement rate limiting
7. Setup backup database

## License

MIT License
