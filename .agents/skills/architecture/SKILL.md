# Skill: Feature-Driven Layered Architecture (Gin + GORM)

This document enforces the structural rules, data flow boundaries, and dependency injection patterns for the `vilamuzz-yota-backend` workspace. AI agents must strictly adhere to these layer separations when creating or modifying features.

---

## 1. Directory Blueprint & Domain Anatomy

Use a **Feature-Driven** module architecture inside the `app/` directory. Each domain feature must remain self-contained.

```text
app/
└── {domain}/
    ├── {domain}.go    # Domain Entities, Core Structs, and Layer Interfaces
    ├── handler.go     # HTTP Transport Layer (Gin Gonic)
    ├── service.go     # Core Business Logic Layer
    ├── repository.go  # Persistence Layer (GORM)
    ├── request.go     # Input DTOs and Validation Rules
    └── response.go    # Output DTOs / Serializers
```

---

## 2. Unidirectional Data Flow & Boundaries

Data flow is completely linear and strict. Skipping layers is an un-mergeable architectural violation.

```text
[HTTP Request] ──> handler.go ──> service.go ──> repository.go ──> [GORM / PostgreSQL]
```

### A. The Handler Layer (`handler.go`)

- **Framework:** Gin Gonic (`*gin.Context`)
- **Responsibilities:** Routing HTTP endpoints, processing parameters, parsing incoming inputs using `request.go` constraints, and serving standardized JSON structures from `response.go`
- **Constraint:** Zero business decisions, calculations, or raw database queries. Execute logic solely by triggering the Service interface.

### B. The Service Layer (`service.go`)

- **Responsibilities:** Validating business rules, managing multi-module business coordination, and orchestrating database transactions
- **Constraint:** Completely decoupled from HTTP components. Passing `*gin.Context` into this layer is strictly prohibited. Accept raw Go types/DTOs and `context.Context` only.

### C. The Repository Layer (`repository.go`)

- **Framework:** GORM (`*gorm.DB`)
- **Responsibilities:** Executing queries, managing structural joins, filters, and persistence mechanics
- **Constraint:** No business rules. If data is missing or a query fails, return clean database errors to the service layer.

---

## 3. Implementation Contracts & Dependency Injection

Preserve compile-time isolation by using interfaces defined within the respective files (e.g., `service.go` and `repository.go`). Concrete structs should never be directly passed between layers.

### 1. Domain Entities (`app/account/account.go`)

```go
package account

import (
	"time"
	"github.com/google/uuid"
)

type Account struct {
	ID        uuid.UUID `gorm:"primaryKey" json:"id"`
	Email     string    `gorm:"unique;not null" json:"email"`
	Password  string    `gorm:"not null" json:"password"`
	IsBanned  bool      `gorm:"default:false" json:"isBanned"`
	// ... other fields and relationships (UserProfile, AccountRoles)
}

type UserProfile struct {
	ID        uuid.UUID `gorm:"primaryKey" json:"id"`
	AccountID uuid.UUID `gorm:"unique;not null" json:"accountId"`
	Username  string    `json:"username"`
	// ... other fields
}

// ... other structs (Role, AccountRole) and Role constants
```

### 2. Interface & Persistence Implementation (`app/account/repository.go`)

```go
package account

import (
	"context"
	"gorm.io/gorm"
)

type Repository interface {
	FindAllAccounts(ctx context.Context, options map[string]interface{}) ([]Account, error)
	FindOneAccount(ctx context.Context, options map[string]interface{}) (*Account, error)
	// ... other repository interface definitions
}

type repository struct {
	Conn *gorm.DB
}

func NewRepository(conn *gorm.DB) Repository {
	return &repository{Conn: conn}
}

func (r *repository) FindAllAccounts(ctx context.Context, options map[string]interface{}) ([]Account, error) {
	// Query build and execution logic goes here (using GORM, preloads, filters, sorting, cursor pagination, etc.)
}

// ... other repository method implementations
```

### 3. Interface & Business Implementation (`app/account/service.go`)

```go
package account

import (
	"context"
	"time"
	"github.com/Vilamuzz/yota-backend/pkg"
)

type Service interface {
	GetAccountList(ctx context.Context, params AccountQueryParam, excludeSuperadmin bool) pkg.Response
	GetAccountByID(ctx context.Context, accountID string) pkg.Response
	// ... other service interface definitions
}

type service struct {
	repo     Repository
	timeout  time.Duration
	s3Client s3_pkg.Client // other client dependencies
}

func NewService(r Repository, timeout time.Duration, s3Client s3_pkg.Client) Service {
	return &service{
		repo:     r,
		timeout:  timeout,
		s3Client: s3Client,
	}
}

func (s *service) GetAccountList(ctx context.Context, params AccountQueryParam, excludeSuperadmin bool) pkg.Response {
	// 1. Business logic/validations, setting up query options
	// 2. Query persistence layer: s.repo.FindAllAccounts(ctx, options)
	// 3. Process and prepare response data (e.g., cursor pagination)
	// 4. Return standard response pkg.NewResponse(http.StatusOK, "message", nil, toAccountListResponse(accounts, ...))
}

// ... other service method implementations
```

### 4. Transport Implementation (`app/account/handler.go`)

```go
package account

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/Vilamuzz/yota-backend/pkg"
	"github.com/Vilamuzz/yota-backend/app/middleware"
)

type handler struct {
	service    Service
	middleware middleware.AppMiddleware
}

func NewHandler(r *gin.RouterGroup, s Service, m middleware.AppMiddleware) {
	handler := &handler{
		service:    s,
		middleware: m,
	}
	handler.RegisterRoutes(r)
}

func (h *handler) RegisterRoutes(r *gin.RouterGroup) {
	// Register endpoints with appropriate middlewares
	r.GET("/roles", h.GetRoleList)

	admin := r.Group("/admin/accounts")
	admin.Use(h.middleware.RequireRoles(...))
	{
		admin.GET("", h.GetAccountList)
		admin.GET("/:accountId", h.GetAccountByID)
	}
}

func (h *handler) GetAccountList(c *gin.Context) {
	ctx := c.Request.Context()
	var queryParam AccountQueryParam

	// 1. Parameter binding using Gin Context query/json binding
	if err := c.ShouldBindQuery(&queryParam); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Permintaan tidak valid", nil, nil))
		return
	}

	// 2. Service execution
	res := h.service.GetAccountList(ctx, queryParam, false)

	// 3. Serve standard json response
	c.JSON(res.Status, res)
}

// ... other route handlers
```

### 5. Input Requests (`app/account/request.go`)

```go
package account

import (
	"github.com/Vilamuzz/yota-backend/pkg"
	"github.com/Vilamuzz/yota-backend/pkg/enum"
)

type UpdateUserProfileRequest struct {
	Username             string `json:"username" form:"username"`
	Email                string `json:"email" form:"email"`
	DefaultAccountRoleID int    `json:"defaultAccountRoleId" form:"defaultAccountRoleId"`
	// ... other fields
}

type AccountQueryParam struct {
	Search    string             `form:"search"`
	RoleID    int                `form:"roleId"`
	IsBanned  *bool              `form:"isBanned"`
	SortOrder enum.SortOrderEnum `form:"sortOrder"`
	pkg.PaginationParams
}
```

### 6. Output Responses (`app/account/response.go`)

```go
package account

import (
	"time"
	"github.com/Vilamuzz/yota-backend/pkg"
)

type AccountResponse struct {
	ID        string                 `json:"id"`
	Username  string                 `json:"username"`
	Email     string                 `json:"email"`
	IsBanned  bool                   `json:"isBanned"`
	Roles     []AccountRolesResponse `json:"roles"`
	CreatedAt time.Time              `json:"createdAt"`
}

func (a *Account) toAccountResponse() AccountResponse {
	// Conversion logic from entity/model to Response DTO goes here
}
```

---

## 4. FORBIDDEN AI ANTI-PATTERNS (DO NOT GENERATE)

- **Context Bleeding:** Do not use or parse `*gin.Context` inside any file matching `*_service.go` or `*_repository.go`
- **Layer Bypass:** Never pass a `*gorm.DB` connection directly into a service or handler method. All persistence queries must reside inside `repository.go`
- **Cross-Module Imports:** Do not import concrete implementation structs between different modules. Inter-module calls must go through the respective module's Service interface via dependency injection wire-ups
- **Validation Separation:** Do not put Gin binding tags (`binding:"required"`) on domain core models. Validation annotations must exist exclusively within `request.go` structs
