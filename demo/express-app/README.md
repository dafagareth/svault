# DevNotes API

REST API for managing developer notes and code snippets, built with **Express.js**, **Prisma**, and **SQLite**.

## Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/auth/register` | — | Register a new account |
| POST | `/auth/login` | — | Login, get JWT token |
| GET | `/notes` | Bearer | List your notes (supports `?search=`, `?language=`, `?pinned=`) |
| POST | `/notes` | Bearer | Create a note |
| GET | `/notes/:id` | Bearer | Get a note |
| PUT | `/notes/:id` | Bearer | Update a note |
| DELETE | `/notes/:id` | Bearer | Delete a note |
| GET | `/health` | — | Health check |

## Setup

### With svault (recommended)

```bash
# 1. Configure secrets — runs once per machine
./setup.sh

# 2. Install dependencies and migrate DB
npm install
npx prisma migrate dev --name init

# 3. Start — secrets injected automatically, never touch disk
svault exec -- npm start
```

### Traditional way (less secure)

```bash
# 1. Copy template and fill in values manually
cp .env.example .env
# edit .env — secrets now live as plaintext on disk

# 2. Install and run
npm install
npx prisma migrate dev --name init
npm start
```

> With svault, secrets never exist as plaintext files. They are stored encrypted in `~/.svault/vault.enc` and injected into the process at runtime.

## Secrets

| Variable | Required | Description |
|----------|----------|-------------|
| `DATABASE_URL` | Yes | SQLite file path, e.g. `file:./dev.db` |
| `JWT_SECRET` | Yes | Secret key for signing JWT tokens |
| `PORT` | No | Server port (default: 3000) |
| `RESEND_API_KEY` | No | Resend API key for welcome emails |
| `EMAIL_FROM` | No | Sender address for emails |
