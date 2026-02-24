# Shoop - Feng Shui E-Commerce Platform

A modular e-commerce web application built with Go (Echo framework), SQLite (GORM), and server-side rendered HTML templates.

## Architecture

The application is split into two independent services sharing the same database:

| Service | Port | Purpose |
|---------|------|---------|
| **Web** (Frontend) | `:8600` | End-user storefront |
| **Admin** (Back-office) | `:18600` | Management dashboard |

### Tech Stack

- **Go 1.23+** with Echo v4 framework
- **SQLite** via GORM (easily swappable to PostgreSQL/MySQL)
- **html/template** with modular layouts (base + partials)
- **Tailwind CSS** (CDN) + Font Awesome icons
- **Gorilla Sessions** for separate admin/web session stores

## Quick Start

### Prerequisites

- Go 1.23+
- GCC (for CGO/SQLite)

### Run Locally

```bash
# Start the frontend (port 8600)
go run ./cmd/web/

# Start the admin panel (port 18600) — in another terminal
go run ./cmd/admin/
```

The database is auto-created and seeded on first run.

### Default Admin Credentials

- **Email:** `admin@occ.io.vn`
- **Password:** `admin123`

### Docker

```bash
docker compose up --build
```

- Frontend: http://localhost:8600
- Admin: http://localhost:18600

## Project Structure

```
├── cmd/
│   ├── admin/main.go          # Admin server entry point
│   └── web/main.go            # Web server entry point
├── config/                    # App configuration
├── database/
│   ├── database.go            # GORM init + migrations
│   └── seeders/               # Seed data
├── internal/
│   ├── models/                # GORM models
│   ├── middleware/             # Auth & session middleware
│   └── handlers/
│       ├── admin/             # Admin controllers
│       └── web/               # Frontend controllers
├── pkg/
│   ├── session/               # Session management
│   └── utils/                 # Template helpers & renderer
├── templates/
│   ├── admin/                 # Admin templates
│   │   ├── layouts/base.html
│   │   ├── partials/
│   │   └── pages/
│   └── web/                   # Frontend templates
│       ├── layouts/base.html
│       ├── partials/
│       └── pages/
├── static/                    # CSS, JS, images
├── uploads/                   # User-uploaded files
├── tests/                     # Test suite
├── Dockerfile
└── docker-compose.yml
```

## Features

### Admin Panel
- Dashboard with statistics
- CRUD: Categories, Products (with image upload), Orders, Users
- Banner management (SEO sliders)
- Company info & About page editor
- 3-color palette: Light Green, Black, White

### Frontend Store
- Responsive Feng Shui themed design
- Product catalog with category filtering & search
- Product detail with image gallery
- Shopping cart with session storage
- Login/Register modal (triggered on "Add to Cart" for anonymous users)
- Checkout flow with order creation
- Banner slider on homepage

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `APP_ENV` | `development` | Environment mode |
| `ADMIN_PORT` | `18600` | Admin server port |
| `WEB_PORT` | `8600` | Web server port |
| `DB_PATH` | `data/shoop.db` | SQLite database path |
| `SESSION_SECRET` | (set in config) | Session encryption key |
| `UPLOAD_DIR` | `uploads` | File upload directory |

## Testing

The project includes comprehensive tests covering unit tests, API/handler tests, and end-to-end tests.

### Run All Tests

```bash
go test ./... -v
```

### Run Specific Test Suites

```bash
# Unit tests (models, helpers)
go test ./tests/unit/... -v

# API/handler tests (admin + web endpoints)
go test ./tests/api/... -v

# End-to-end tests (full user flows)
go test ./tests/e2e/... -v
```

### Test Coverage

```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Database Migration to PostgreSQL/MySQL

The app uses GORM, so switching databases requires only changing the driver:

```go
// In database/database.go, replace:
import "gorm.io/driver/sqlite"
// With:
import "gorm.io/driver/postgres"
// And update the Open() call accordingly.
```

## License

MIT
