## Web app Implemantation Template

This is a template repository for implementing web applications using Go and Next.js.

### Features

- Backend: Go (Echo)
- Frontend: Next.js (React)
- Database: MySQL
- ORM: GORM
- Authentication: NextAuth(Google)

### Prerequisites

- Go 1.25
- Node.js 18+
- MySQL 8+
- pnpm 10+
- Docker
- Docker Compose

### Setup Instructions

1. Clone the repository:

   ```bash
   git clone  <repository_url>
   cd go-template-api
   ``` 
2. Set up the backend:

   ```bash
   cd backend

   # Start MySQL + the API server (hot reload) with Docker Compose.
   # Dev credentials/ports are defined in backend/compose.yml.
   docker compose up --build

   # In another shell, run DB migrations against the MySQL container
   # (connecting from the host, so use localhost instead of the "db" service name)
   export DATABASE_URL="user:P@ssw0rd@tcp(localhost:3306)/test?parseTime=true"
   go run cmd/migrate/main.go -command up
   go run cmd/migrate/main.go -command seed   # optional: load sample data
   ```

   The API server listens on `http://localhost:8080`.

3. Set up the frontend:

   ```bash
   cd frontend
   # Create a .env file with at least:
   #   NEXT_PUBLIC_API_BASE_URL=http://localhost:8080
   #   GOOGLE_CLIENT_ID=<your-google-oauth-client-id>
   #   GOOGLE_CLIENT_SECRET=<your-google-oauth-client-secret>
   #   NEXTAUTH_SECRET=<random-secret>

   pnpm install
   pnpm dev
   ```

4. Access the application at `http://localhost:3000`.
5. Stop the backend:

   ```bash
   cd backend
   docker compose down
   ```

### Configuration

- Local backend dev credentials/ports are defined directly in `backend/compose.yml`. For non-Docker or deployed environments, set `ENV`, `PORT`, `DATABASE_URL`, `PROJECTID` (see `pkg/config/config.go`).
- Customize authentication providers (Google OAuth client ID/secret, `NEXTAUTH_SECRET`) in `frontend/.env`.
- Coding conventions for this template live under `.claude/rules/` (architecture, error handling, testing, etc.) and `.github/instructions/`.
- `.claude/` is the source of truth for skills, hooks, and agents; `.github/skills` and `.agents` are symlinks into it.
- Feel free to contribute to this template by submitting issues or pull requests.
- Happy coding!

### Documentation for AI agents

- [`CLAUDE.md`](CLAUDE.md) is the entry point for AI coding agents (Claude Code and other `AGENTS.md`-aware tools). `AGENTS.md` is a symlink to it.
- [`DESIGN.md`](DESIGN.md) documents the `frontend/` UI design (design system tokens and the EC sample's screen specs). Backend architecture decisions and the reasoning behind them live in `CLAUDE.md`'s "アーキテクチャの設計判断" section.
- `.claude/skills/` holds task-specific workflows: `implement-api`, `implement-component`, `migration-db-schema`, `design-feature` (feature design), plus `commit`, `create-pr`, and `code-review` for the git/GitHub workflow.

### CI

- `.github/workflows/lint_and_test.yml`: backend tests and `golangci-lint`.
- `.github/workflows/security.yml`: secret scanning (gitleaks, trufflehog), SAST (Semgrep), and dependency vulnerability checks (`govulncheck`, `pnpm audit`).
- `.github/workflows/deploy-cloudrun.yml` / `deploy-ecs.yml`: deployment.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details
