# Skill: Stack Environment & Tool Version Enforcement

This document enforces version compliance for the `vilamuzz-yota-backend` runtime and storage engine. AI tools must strictly generate code that conforms to the language versions and declarative migration strategies specified below.

---

## 1. Precise Core Stack Baselines

| Component       | Target Version | Compliance & Code Constraints                                               |
| :-------------- | :------------- | :-------------------------------------------------------------------------- |
| **Go (Golang)** | `1.25.0`       | Utilize modern core library structures. Never use deprecated packages.      |
| **Gin Gonic**   | `v1.11.0`      | Standardized JSON payload handling via Context Bind utilities.              |
| **GORM**        | `v1.31.1`      | Strictly object-relational mapping patterns; no untyped raw string queries. |
| **PostgreSQL**  | `16`           | Leverage modern engine optimizations, JSONB filtering, and index types.     |
| **Atlas CLI**   | `Declarative`  | Managed exclusively via HCL/Schema synchronization via `atlas.hcl`.         |

---

## 2. Advanced Go 1.25+ Syntax Rules

The AI tool must write contemporary Go 1.25+ code. Legacy patterns will trigger CI lint pipeline errors.

### A. I/O Operations (Strictly Enforced)

The `io/ioutil` package is completely forbidden. Use native alternatives:

- Use `io.ReadAll` instead of `ioutil.ReadAll`
- Use `os.ReadFile` instead of `ioutil.ReadFile`
- Use `os.WriteFile` instead of `ioutil.WriteFile`

### B. Structured Logging (Logrus)

Structured logging in this project must leverage the `github.com/sirupsen/logrus` package:

```go
import "github.com/sirupsen/logrus"

// Correct structured logging usage
logrus.WithFields(logrus.Fields{
	"component":      "donation.service",
	"transaction_id": txID,
}).Info("executing transaction processing")
```

---

## 3. Database & Atlas Declarative Migration Workflow

We use the Atlas Declarative Workflow via `atlas.hcl` to sync database status.

### A. Migration Restrictions

- **No Manual SQL Files**: AI must never suggest writing or appending manual SQL alterations directly inside the `/migrations` folder.
- **Model-Driven Changes**: Schema updates must be performed solely by modifying structural definitions inside the model files (e.g. `app/{domain}/account.go`, etc.) and registered in `models/postgre.go`. Atlas computes the declarative diff against the live PostgreSQL 16 instance.

### B. GORM Structure Definitions (PostgreSQL 16 Target)

When creating or expanding database model layers, explicitly define exact column constraints via structural tags:

```go
type DonationProgram struct {
	ID          uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Title       string         `gorm:"type:varchar(255);not null" json:"title"`
	Description string         `gorm:"type:text" json:"description"`
	Metadata    map[string]any `gorm:"type:jsonb" json:"metadata"` // Native Postgres JSONB compliance
	CreatedAt   time.Time      `gorm:"not null" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"not null" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
```

---

## 4. FORBIDDEN AI ANTI-PATTERNS (DO NOT GENERATE)

- **Legacy Package Usage**: Do not include `io/ioutil` or `os.SEEK_SET` references.
- **Raw Migration Writing**: Do not write manual `ALTER TABLE` or `CREATE TABLE` raw migration files inside `/migrations/`. Ensure schema variations are built cleanly as Go structs for Atlas processing.
- **Outdated Database Assertions**: Do not assume structural engine capabilities from old PostgreSQL configurations. Generate modern queries making optimal use of v16 performance standards.
