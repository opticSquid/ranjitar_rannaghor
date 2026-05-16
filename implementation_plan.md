# Add Test Infrastructure with Database Mocking & Integration Tests

## Background

The codebase currently has only **one test file** (`journal/handler_test.go`) which tests the pure `CalculateTotalCost` function. All HTTP handlers directly call `database.GetDbConn()` — a package-level singleton — making them **impossible to unit test** without a real database.

This plan adds:
1. A **testcontainers-based integration test infrastructure** using a real PostgreSQL instance
2. A **refactored `database` package** that supports dependency injection for testing
3. **Comprehensive handler tests** for all 7 packages (users, wallet, journal, billing, expenses, meals, stats)

---

## User Review Required

> [!IMPORTANT]
> **Minimal refactoring approach**: Rather than introducing interfaces for every database interaction (which would be a massive rewrite), I'm proposing a lightweight approach: make `database.GetDbConn()` configurable via a `SetDbConn()` function. This lets integration tests inject a test database pool with **zero changes** to any handler code. All handlers continue calling `database.GetDbConn()` as before.

>>> [USER COMMENT] Do it but remember it should not break existing code and logic

> [!WARNING]
> **Testcontainers requires Docker**: The integration tests use [testcontainers-go](https://github.com/testcontainers/testcontainers-go) to spin up a real PostgreSQL container. Docker must be available in the CI/test environment. If Docker is not available, these tests will be skipped via a build tag.

>>> [USER COMMENT] Yes the system has docker installed on it

---

## Open Questions

1. **Custom enum types**: The schema uses custom PostgreSQL types (`shift`, `user_type`, `subscription_type`, `txn_type`, `txn_status`). I'll create these in the test database setup. Do you have preferred enum values, or should I infer them from the codebase usage? (I'll infer: `shift` = `lunch|dinner`, `user_type` = `normal|admin`, `subscription_type` = `standard|special`, `txn_type` = `recharge|delivery|refund`, `txn_status` = `confirmed|pending_acknowledgement`)

>>> [USER COMMENT] You decide which would be better for long term codebase stability and do the required task.
---

## Proposed Changes

### Test Infrastructure (`testdb` package)

#### [NEW] [testdb.go](file:///home/soumalya/Development/ranjitar_rannaghor/business-apps/admin-and-billing/testdb/testdb.go)

A shared test helper package that:
- Starts a PostgreSQL container via `testcontainers-go`
- Runs the full DDL schema (including custom enum types)
- Seeds default meal prices
- Provides `Setup()` / `Teardown()` / `ResetData()` functions
- Injects the test pool via `database.SetDbConn()`

---

### Database Package Refactoring

#### [MODIFY] [db.go](file:///home/soumalya/Development/ranjitar_rannaghor/business-apps/admin-and-billing/database/db.go)

Add a single exported function:
```go
// SetDbConn allows tests to inject a test database pool.
func SetDbConn(pool *pgxpool.Pool) {
    dbPool = pool
}
```

This is the **only production code change**. All handlers continue calling `GetDbConn()` unchanged.

---

### Integration Tests — By Package

Each test file follows the same pattern:
1. `TestMain(m *testing.M)` calls `testdb.Setup()` / `testdb.Teardown()`
2. Each test calls `testdb.ResetData()` to get a clean slate
3. Tests construct `httptest.NewRequest` + `httptest.NewRecorder` and call handler functions directly
4. Assertions verify HTTP status codes and JSON response bodies

---

#### [NEW] [handlers_test.go](file:///home/soumalya/Development/ranjitar_rannaghor/business-apps/admin-and-billing/users/handlers_test.go)

| Test | What it verifies |
|------|-----------------|
| `TestCreateUser_Success` | Creates a user, verifies 200 + returned user_id |
| `TestCreateUser_InvalidJSON` | Malformed body → 400 |
| `TestCreateUser_DefaultRole` | Empty role defaults to "normal" |
| `TestGetUsers_Empty` | No users → empty JSON array |
| `TestGetUsers_WithBalance` | User with wallet transactions → correct balance via `DISTINCT ON` query |

---

#### [NEW] [handler_test.go](file:///home/soumalya/Development/ranjitar_rannaghor/business-apps/admin-and-billing/wallet/handler_test.go)

| Test | What it verifies |
|------|-----------------|
| `TestRechargeWallet_Success` | Recharge creates a wallet_transaction with correct balance_after |
| `TestRechargeWallet_InvalidJSON` | Malformed body → 400 |
| `TestRechargeWallet_MultipleRecharges` | Sequential recharges accumulate balance correctly |
| `TestRechargeWallet_BackdatedRecharge` | Recharge with past date triggers balance recalculation |

---

#### [NEW] [handler_integration_test.go](file:///home/soumalya/Development/ranjitar_rannaghor/business-apps/admin-and-billing/journal/handler_integration_test.go)

| Test | What it verifies |
|------|-----------------|
| `TestCreateDailyEntry_Success` | Creates entry + wallet delivery transaction, verifies new_balance |
| `TestCreateDailyEntry_DeductsFromWallet` | After recharge, entry deducts from balance correctly |
| `TestDeleteDailyEntry_Refund` | Deleting entry creates refund transaction, restores balance |
| `TestUpdateDailyEntry_CostDiff` | Updating entry to higher cost creates additional delivery txn |
| `TestUpdateDailyEntry_NoCostChange` | Same cost → no new wallet transaction |
| `TestGetDailyEntries_ByDate` | Filters entries by date correctly |
| `TestGetDailyEntries_ByUser` | Filters entries by user_id correctly |
| `TestGetDailyEntries_MissingDate` | Missing date param → 400 |

---

#### [NEW] [handler_test.go](file:///home/soumalya/Development/ranjitar_rannaghor/business-apps/admin-and-billing/billing/handler_test.go)

| Test | What it verifies |
|------|-----------------|
| `TestGetBill_FullReport` | Bill report with logs, opening/closing balance, total recharges |
| `TestGetBill_EmptyPeriod` | No logs in period → empty logs, zero totals |
| `TestGetBill_OpeningBalance` | Opening balance derived from last txn before start_date |

---

#### [NEW] [handler_test.go](file:///home/soumalya/Development/ranjitar_rannaghor/business-apps/admin-and-billing/expenses/handler_test.go)

| Test | What it verifies |
|------|-----------------|
| `TestCreateExpense_Success` | Creates expense, verifies returned expense_id |
| `TestGetExpenses_All` | Lists all expenses |
| `TestGetExpenses_DateRange` | Filters by start_date/end_date |
| `TestUpdateExpense_Success` | Updates expense fields |
| `TestDeleteExpense_Success` | Deletes expense, verifies it's gone |

---

#### [NEW] [handler_test.go](file:///home/soumalya/Development/ranjitar_rannaghor/business-apps/admin-and-billing/meals/handler_test.go)

| Test | What it verifies |
|------|-----------------|
| `TestGetMeals_DefaultPrices` | Returns seeded default meal prices |
| `TestUpdateMeal_Success` | Updates a meal price |
| `TestGetMealPricesInternal` | Internal function returns correct map |

---

#### [NEW] [handler_test.go](file:///home/soumalya/Development/ranjitar_rannaghor/business-apps/admin-and-billing/stats/handler_test.go)

| Test | What it verifies |
|------|-----------------|
| `TestGetDashboardStats` | Revenue, expenses, profit, active customers, wallet pool |
| `TestGetAnalyticsStats_Trends` | 30-day trend data populated correctly |
| `TestGetAnalyticsStats_MealTypeDistribution` | Standard vs special counts |
| `TestGetAnalyticsStats_ShiftDistribution` | Lunch vs dinner counts |

---

### Utils Tests

#### [NEW] [wallet_balance_test.go](file:///home/soumalya/Development/ranjitar_rannaghor/business-apps/admin-and-billing/utils/wallet_balance_test.go)

| Test | What it verifies |
|------|-----------------|
| `TestRecalculateBalances_SingleTxn` | Single transaction recalculated correctly |
| `TestRecalculateBalances_MixedTypes` | Delivery deducts, recharge adds |
| `TestRecalculateBalances_NoTransactions` | No txns after fromTime → no-op |
| `TestRecalculateBalances_ChainedBalances` | Multiple txns produce correct running balance |

---

## New Dependencies

| Dependency | Purpose |
|------------|---------|
| `github.com/testcontainers/testcontainers-go` | Spin up PostgreSQL containers for integration tests |
| `github.com/testcontainers/testcontainers-go/modules/postgres` | PostgreSQL-specific container module |
| `github.com/stretchr/testify` | Assertion helpers (`assert`, `require`) |

---

## Verification Plan

### Automated Tests
```bash
# Run all tests (requires Docker)
cd business-apps/admin-and-billing && go test ./... -v -count=1

# Run only unit tests (no Docker needed)
cd business-apps/admin-and-billing && go test ./journal/ -v -run TestCalculateTotalCost

# Run specific integration package
cd business-apps/admin-and-billing && go test ./wallet/ -v -count=1
```

### Manual Verification
- Confirm all tests pass with `go test ./... -v`
- Verify existing `TestCalculateTotalCost` still passes
- Ensure `go build` succeeds (no production code broken)
