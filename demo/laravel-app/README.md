# QuickTask API

Task management REST API built with **Laravel 11** and **Sanctum** token authentication.

## Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/api/register` | — | Register a new account |
| POST | `/api/login` | — | Login, get Sanctum token |
| POST | `/api/logout` | Bearer | Logout |
| GET | `/api/me` | Bearer | Get current user |
| GET | `/api/tasks` | Bearer | List tasks (`?status=`, `?priority=`, `?search=`) |
| POST | `/api/tasks` | Bearer | Create a task |
| GET | `/api/tasks/{id}` | Bearer | Get a task |
| PUT | `/api/tasks/{id}` | Bearer | Update a task |
| DELETE | `/api/tasks/{id}` | Bearer | Delete a task |

## Setup

### With svault (recommended)

```bash
# 1. Configure secrets — runs once per machine
./setup.sh

# 2. Install PHP dependencies
composer install

# 3. Run migrations
svault exec -- php artisan migrate

# 4. Start the server — secrets injected automatically, never touch disk
svault exec -- php artisan serve
```

### Traditional way (less secure)

```bash
# 1. Copy template and fill in values manually
cp .env.example .env
# edit .env — DB credentials, APP_KEY, MAIL config now live as plaintext

# 2. Install and run
composer install
php artisan key:generate
php artisan migrate
php artisan serve
```

> With svault, secrets never exist as plaintext files. They are stored encrypted in `~/.svault/vault.enc` and injected into the process at runtime.

## Secrets

| Variable | Required | Description |
|----------|----------|-------------|
| `APP_KEY` | Yes | Laravel application key (base64 encoded) |
| `APP_ENV` | Yes | Environment (`local`, `production`) |
| `DB_HOST` | Yes | MySQL/PostgreSQL host |
| `DB_DATABASE` | Yes | Database name |
| `DB_USERNAME` | Yes | Database user |
| `DB_PASSWORD` | Yes | Database password |
| `MAIL_HOST` | No | SMTP host for email notifications |
| `MAIL_USERNAME` | No | SMTP username |
| `MAIL_PASSWORD` | No | SMTP password |

## Task Schema

```json
{
  "id": 1,
  "title": "Implement authentication",
  "description": "Add JWT-based login flow",
  "status": "in_progress",
  "priority": "high",
  "due_date": "2025-01-31",
  "user_id": 1,
  "created_at": "2025-01-01T00:00:00Z",
  "updated_at": "2025-01-01T00:00:00Z"
}
```
