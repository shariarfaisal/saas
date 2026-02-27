//go:build ignore

// seed_dev.go seeds the development database with a single dev tenant and an admin user.
// Run with: go run ./scripts/seed_dev.go
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/munchies/platform/backend/internal/config"
	"github.com/munchies/platform/backend/internal/db/sqlc"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, cfg.Database.URL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect db: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	q := sqlc.New(pool)

	// ─── Dev tenant ──────────────────────────────────────────────────────────
	settings, _ := json.Marshal(map[string]interface{}{
		"allow_cod":           true,
		"min_order_amount":    0,
		"free_delivery_above": 0,
	})

	commissionRate := pgtype.Numeric{}
	if err := commissionRate.Scan("10.00"); err != nil {
		fmt.Fprintf(os.Stderr, "parse commission rate: %v\n", err)
		os.Exit(1)
	}

	tenant, err := q.CreateTenant(ctx, sqlc.CreateTenantParams{
		Slug:           "dev",
		Name:           "Dev Tenant",
		Status:         sqlc.TenantStatusActive,
		Plan:           sqlc.TenantPlanStarter,
		CommissionRate: commissionRate,
		Settings:       json.RawMessage(settings),
		ContactEmail:   "dev@example.com",
		ContactPhone:   sql.NullString{},
		Address:        nil,
		Timezone:       "Asia/Dhaka",
		Currency:       "BDT",
		Locale:         "en",
		PrimaryColor:   "#FF6B35",
		SecondaryColor: "#2C3E50",
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "create tenant: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Tenant created: id=%s slug=%s\n", tenant.ID, tenant.Slug)

	// ─── Admin user ──────────────────────────────────────────────────────────
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		fmt.Fprintf(os.Stderr, "hash password: %v\n", err)
		os.Exit(1)
	}

	tenantPgUUID := pgtype.UUID{Bytes: tenant.ID, Valid: true}

	adminUser, err := q.CreateUser(ctx, sqlc.CreateUserParams{
		TenantID:     tenantPgUUID,
		Name:         "Dev Admin",
		Email:        sql.NullString{String: "admin@dev.example.com", Valid: true},
		PasswordHash: sql.NullString{String: string(passwordHash), Valid: true},
		Role:         sqlc.UserRoleTenantAdmin,
		Status:       sqlc.UserStatusActive,
		Metadata:     json.RawMessage("{}"),
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "create admin user: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Admin user created: id=%s email=admin@dev.example.com password=password123\n", adminUser.ID)
	fmt.Println()
	fmt.Println("Dev seed complete. Use the credentials above to log in.")
}
