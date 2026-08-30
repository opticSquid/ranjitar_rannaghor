# Copilot instructions for this repo

- This repo is a small monorepo: the business API is in `business-apps/admin-and-billing`, and the web UI is in `business-apps/admin-and-billing/frontend`.
- The Go app is a single `chi` server started in `business-apps/admin-and-billing/main.go`. All routes are mounted under `/api` and grouped by domain (`users`, `wallet`, `journal`, `meals`, `expenses`, `billing`, `stats`).
- Follow the existing package pattern when adding features: `models.go` for request/response structs, `repository.go` for DB queries via `database.GetDbConn()`, `services.go` for business logic, and `controllers.go` for HTTP handlers. Example: `meals/` and `journal/` follow this layout.
- Database access is centralized in `business-apps/admin-and-billing/database/db.go`. `InitDB()` reads `POSTGRES_*` env vars, creates bootstrapped tables if missing, and seeds default meal prices. Reuse this instead of opening ad hoc DB connections.
- The backend uses `pgx` and PostgreSQL. Many modules query tables directly and return JSON; there is no repository interface layer or service container pattern.
- The frontend is SolidJS + Vite. Pages live under `business-apps/admin-and-billing/frontend/src/pages`, shared app state under `src/store`, and API calls are direct `axios` requests to `/api/*`.
- Keep UI logic in the same style as `frontend/src/pages/DailyEntry.tsx`: fetch data on mount/effect, update local state with `createSignal`, and mutate global caches using helpers like `updateUserBalance()` in `src/store/userStore.ts`.
- Price history is a real domain concept here: `meals/repository.go` stores effective prices in `meal_price_history` and resolves “price as of timestamp” via `GetPriceAtForItem()` / `GetPricesAt()`. Tests in `meals/price_boundary_integration_test.go` cover boundary conditions.
- For tests, prefer the repo’s established PostgreSQL Testcontainers setup in `business-apps/admin-and-billing/testdb/testdb.go`; tests usually call `testdb.Setup()` / `ResetData()` rather than assuming a local DB is already running.
- Standard backend verification: `cd business-apps/admin-and-billing && go test ./...`
- Frontend build: `cd business-apps/admin-and-billing/frontend && npm install && npm run build`
- Local stack: use the root `docker-compose.yaml` / `compose.yaml` for the Postgres + app services; `MODE` is either `dev` or `prod`, and `DOMAIN` is required in production mode.
- Production mode in `main.go` switches to HTTPS/HTTP3 with ACME and serves the built frontend from `dist/`; dev mode serves plain HTTP and still serves `index.html` for client-side routes.
- Keep naming consistent with existing code: `GetX`, `CreateX`, `UpdateX`, `DeleteX` for handlers and service methods; return JSON from handlers and set explicit HTTP status codes.
- When changing APIs, update both the Go route registrations in `main.go` and the frontend callers that hit those endpoints; this repo does not have generated clients or a separate API schema.
- Avoid introducing new frameworks or abstraction layers unless the module already uses them; the project is intentionally direct and domain-by-domain rather than layered microservice-style.
- For pricing or billing logic, check the downstream use in `journal` and `billing` packages before changing calculation semantics, because these modules are tightly tied to the meal-price data model.
- If you add a new domain package, mimic the existing folder layout and ensure it is included in the root `main.go` router and any relevant tests.
