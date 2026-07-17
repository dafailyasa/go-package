package healthcheck

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// Status represents the health status
type Status string

const (
	StatusUp   Status = "UP"
	StatusDown Status = "DOWN"
)

// Response is the health check response
type Response struct {
	Status    Status            `json:"status"`
	Service   string            `json:"service"`
	Timestamp time.Time         `json:"timestamp"`
	Checks    map[string]Check  `json:"checks,omitempty"`
}

// Check represents an individual health check
type Check struct {
	Status  Status `json:"status"`
	Message string `json:"message,omitempty"`
}

// HealthChecker provides health check functionality
type HealthChecker struct {
	serviceName string
	db          *sql.DB
	checks      map[string]CheckFunc
}

// CheckFunc is a function that performs a health check
type CheckFunc func(ctx context.Context) Check

// NewHealthChecker creates a new health checker
func NewHealthChecker(serviceName string) *HealthChecker {
	return &HealthChecker{
		serviceName: serviceName,
		checks:      make(map[string]CheckFunc),
	}
}

// WithDatabase adds a database health check
func (h *HealthChecker) WithDatabase(db *sql.DB) *HealthChecker {
	h.db = db
	h.checks["database"] = h.checkDatabase
	return h
}

// WithCheck adds a custom health check
func (h *HealthChecker) WithCheck(name string, check CheckFunc) *HealthChecker {
	h.checks[name] = check
	return h
}

// checkDatabase checks database connectivity
func (h *HealthChecker) checkDatabase(ctx context.Context) Check {
	if h.db == nil {
		return Check{Status: StatusDown, Message: "database not configured"}
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := h.db.PingContext(ctx); err != nil {
		return Check{Status: StatusDown, Message: err.Error()}
	}

	return Check{Status: StatusUp}
}

// Check performs all health checks
func (h *HealthChecker) Check(ctx context.Context) Response {
	checks := make(map[string]Check)
	overallStatus := StatusUp

	for name, checkFn := range h.checks {
		check := checkFn(ctx)
		checks[name] = check
		if check.Status == StatusDown {
			overallStatus = StatusDown
		}
	}

	return Response{
		Status:    overallStatus,
		Service:   h.serviceName,
		Timestamp: time.Now().UTC(),
		Checks:    checks,
	}
}

// RegisterRoutes registers health check routes on an Echo group
func (h *HealthChecker) RegisterRoutes(group *echo.Group) {
	// Liveness probe - always returns 200 if process is alive
	group.GET("/ping", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "pong",
		})
	})

	// Readiness probe - checks dependencies
	group.GET("/health", func(c echo.Context) error {
		response := h.Check(c.Request().Context())

		statusCode := http.StatusOK
		if response.Status == StatusDown {
			statusCode = http.StatusServiceUnavailable
		}

		return c.JSON(statusCode, response)
	})
}
