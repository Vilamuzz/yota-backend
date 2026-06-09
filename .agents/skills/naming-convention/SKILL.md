# Skill: Naming Conventions & Project Directory Layout

This document enforces absolute uniformity for naming files, folder structures, database variables, initialisms, and Gin HTTP endpoints within the `vilamuzz-yota-backend` workspace. AI tools must match these structural patterns exactly to ensure seamless code reviews.

---

## 1. Casing & Initialism Standards

We strictly follow Go-idiomatic linting specifications and explicit database mapping formats.

| Target Component | Casing Style | Examples |
| :--- | :--- | :--- |
| **Directory / Folders** | Lowercase `snake_case` | `donation_program`, `foster_children_candidate` |
| **Go Code Files** | Lowercase `snake_case` | `{domain}.go` (e.g. `donation_program.go`), `handler.go`, `service.go` |
| **Go Structs & Interfaces** | `PascalCase` | `DonationProgramTransaction`, `Repository` |
| **Go Local Variables** | `camelCase` | `accountID`, `totalAmount`, `isVerified` |
| **Go Public / Exported Fields**| `PascalCase` | `FindAllAccounts`, `Conn *gorm.DB` |
| **Gin Endpoint URLs** | Lowercase `kebab-case` & Plural | `/api/donation-programs`, `/api/foster-children-candidates` |

### The Go-Idiomatic Acronym Rule (Strict)
Acronyms, initialisms, and abbreviations must be **fully capitalized** or **fully lowercased** based on visibility. Never use mixed casing for initialisms like `Id`, `Http`, or `Jwt`.
* **Correct (Public):** `AccountID`, `JWTAuthMiddleware`, `JSONPayload`
* **Incorrect (Public):** `AccountId`, `JwtAuthMiddleware`, `JsonPayload`
* **Correct (Private):** `accountID`, `jwtToken`, `userID`
* **Incorrect (Private):** `accountId`, `jwtToken`, `userId`

---

## 2. Directory Layout & File Suffixes

Every functional domain module inside the `app/` folder must contain precisely matched file suffixes. Do not deviate from these architectural component mappings:

```text
app/donation_program_expense/
├── donation_program_expense.go # Core models, definitions, and interfaces
├── handler.go                  # Gin routing controllers
├── service.go                  # Business domain logic
├── repository.go               # GORM database queries
├── request.go                  # Incoming DTO structural binding
└── response.go                 # Outgoing API serialization layouts
```

---

## 3. Database Layer & GORM Table Naming

We use GORM's default table namer. GORM automatically converts struct names into snake_case and pluralizes the final noun phrase.

When creating structs inside the `{domain}.go` or model files, the fields must use Go-idiomatic IDs to ensure correct relational database mapping.

Example Mapping Blueprint:
```go
package donation_program

import (
	"time"
	"gorm.io/gorm"
)

// Struct Name: DonationProgramExpense (PascalCase, Singular)
// GORM Default Table Name Generated: donation_program_expenses (snake_case, Plural)
type DonationProgramExpense struct {
	ID                uint           `gorm:"primaryKey" json:"id"`
	DonationProgramID uint           `gorm:"not null" json:"donation_program_id"` // Matches Go-idiomatic initialism
	ExpenseAmount     float64        `gorm:"type:numeric(15,2);not null" json:"expense_amount"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}
```

---

## 4. Gin HTTP Routing Specifications

When declaring API group maps inside handlers or routing wire-ups, routes must follow kebab-case and plural naming conventions.

Routing Blueprint:
```go
package donation_program

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.RouterGroup, h *Handler) {
	// API routes must use plural kebab-case endpoints matching the resource name
	donationGroup := r.Group("/donation-programs")
	{
		donationGroup.POST("", h.CreateDonationProgram)
		donationGroup.GET("", h.GetDonationProgramList)
		donationGroup.GET("/:id", h.GetDonationProgramByID)
	}
}
```

---

## 5. FORBIDDEN AI ANTI-PATTERNS (DO NOT GENERATE)

- **Mixed-Case Acronyms**: Do not generate variables containing `Id`, `Url`, or `Http`. They must be consistently output as `ID`, `URL`, or `HTTP`.
- **Snake-Case Endpoint Targets**: Do not wire paths containing underscores (e.g., `/api/donation_program`). Use kebab-case (e.g., `/api/donation-programs`).
- **Singular Collection URLs**: Do not build collection endpoints with singular resource names (e.g., `/api/donation-program`). Keep collection/resource names pluralized.
- **Custom Table Overrides**: Do not include a `TableName()` method on GORM models unless explicitly ordered by a human controller. Rely completely on the default structural namer.