# Configuration Management dengan Viper

Package ini menggunakan Viper untuk mengelola konfigurasi aplikasi dari file `.env`.

## Fitur

- Load konfigurasi dari file `.env`
- Type-safe config struct
- Default values
- Auto-reload configuration
- Watch file changes
- Global config access

## Cara Penggunaan

### 1. Load Configuration

```go
package main

import (
    "log"
    "service-golang/configs"
)

func main() {
    // Load config dari direktori saat ini
    config, err := configs.LoadConfig(".")
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    // Akses config
    println(config.AppName)
    println(config.ServerPort)
}
```

### 2. Menggunakan Global Config

```go
package main

import (
    "fmt"
    "service-golang/configs"
)

func main() {
    // Load config terlebih dahulu
    configs.LoadConfig(".")

    // Akses dari manapun menggunakan Get()
    cfg := configs.Get()
    fmt.Printf("App: %s, Port: %d\n", cfg.AppName, cfg.ServerPort)
}
```

### 3. Menggunakan Helper Functions

```go
package main

import (
    "service-golang/configs"
)

func main() {
    configs.LoadConfig(".")

    // Akses langsung dengan key
    appName := configs.GetString("APP_NAME")
    port := configs.GetInt("SERVER_PORT")
    debug := configs.GetBool("APP_DEBUG")
}
```

### 4. Watch Configuration Changes

```go
package main

import (
    "service-golang/configs"
)

func main() {
    configs.LoadConfig(".")

    // Enable auto-reload ketika .env berubah
    configs.Watch()

    // Server akan otomatis reload config jika .env diubah
    // ...
}
```

### 5. Manual Reload

```go
package main

import (
    "service-golang/configs"
)

func reloadConfig() {
    if err := configs.Reload(); err != nil {
        log.Printf("Failed to reload: %v", err)
    }
}
```

## Struktur Config

```go
type Config struct {
    // Server Configuration
    ServerHost string
    ServerPort int

    // Database Configuration
    DBHost     string
    DBPort     int
    DBUser     string
    DBPassword string
    DBName     string

    // API Configuration
    APIKey    string
    APISecret string
    BaseURL   string

    // Application Configuration
    AppName        string
    AppEnvironment string
    AppDebug       bool
}
```

## File .env

```env
# Application Configuration
APP_NAME=Service Golang
APP_ENV=development
APP_DEBUG=true

# Server Configuration
SERVER_HOST=localhost
SERVER_PORT=8080

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password_here
DB_NAME=myapp_db

# API Configuration
API_KEY=your-api-key-here
API_SECRET=your-api-secret-here
BASE_URL=http://localhost:8080
```

## Default Values

Jika tidak ada nilai di `.env`, maka akan menggunakan default:

- SERVER_HOST: `localhost`
- SERVER_PORT: `8080`
- DB_PORT: `5432`
- APP_NAME: `Service Golang`
- APP_ENV: `development`
- APP_DEBUG: `true`

## Tips

1. **Jangan commit file .env ke repository**
   - Tambahkan `.env` ke `.gitignore`
   - Gunakan `.env.example` sebagai template

2. **Environment Variables Priority**
   - Viper akan membaca dari environment variables jika ada
   - File .env sebagai fallback

3. **Multiple Environment**
   ```go
   // Development
   configs.LoadConfig(".")

   // Production
   configs.LoadConfig("/etc/myapp")
   ```

## Contoh Lengkap

Lihat implementasi di `cmd/web/server.go` untuk contoh penggunaan lengkap.
