package integration

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	postgresadapter "MikrOps/internal/adapter/outbound/postgres"
	"MikrOps/internal/model"
	contextutil "MikrOps/utils/context"
)

var dsn string

func init() {
	dsn = os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		dsn = "host=localhost port=5432 user=admin password=r00t dbname=mikrotik_billing_test sslmode=disable"
	}
}

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err, "failed to connect to test database")

	err = db.AutoMigrate(&model.User{}, &model.Tenant{}, &model.Customer{})
	require.NoError(t, err, "failed to migrate test database")

	return db
}

func cleanupTestDB(t *testing.T, db *gorm.DB) {
	db.Exec("DELETE FROM users")
	db.Exec("DELETE FROM tenants")
	db.Exec("DELETE FROM customers")
}

func stringPtr(s string) *string {
	return &s
}

func TestTenantIsolation_UserByEmail(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	// Create two tenants
	tenant1 := model.Tenant{
		ID:     uuid.New().String(),
		Name:   "Tenant 1",
		Slug:   stringPtr("tenant1"),
		Status: "active",
	}
	tenant2 := model.Tenant{
		ID:     uuid.New().String(),
		Name:   "Tenant 2",
		Slug:   stringPtr("tenant2"),
		Status: "active",
	}
	db.Create(&tenant1)
	db.Create(&tenant2)

	// Create users with same email for different tenants
	user1 := model.User{
		ID:       uuid.New().String(),
		TenantID: &tenant1.ID,
		Username: "user1",
		Email:    "test@example.com",
		Fullname: "User 1",
	}
	user2 := model.User{
		ID:       uuid.New().String(),
		TenantID: &tenant2.ID,
		Username: "user2",
		Email:    "test@example.com",
		Fullname: "User 2",
	}
	db.Create(&user1)
	db.Create(&user2)

	authAdapter := postgresadapter.NewAuthAdapter(db)

	// Test: Tenant 1 should only find their user
	ctx1 := contextutil.SetTenantID(context.Background(), uuid.MustParse(tenant1.ID))
	foundUser1, err := authAdapter.FindUserByEmail(ctx1, "test@example.com")
	require.NoError(t, err)
	assert.NotNil(t, foundUser1)
	assert.Equal(t, user1.ID, foundUser1.ID)

	// Test: Tenant 2 should only find their user (same email)
	ctx2 := contextutil.SetTenantID(context.Background(), uuid.MustParse(tenant2.ID))
	foundUser2, err := authAdapter.FindUserByEmail(ctx2, "test@example.com")
	require.NoError(t, err)
	assert.NotNil(t, foundUser2)
	assert.Equal(t, user2.ID, foundUser2.ID)

	// Test: Users should not be the same
	assert.NotEqual(t, user1.ID, user2.ID)
}

func TestTenantIsolation_NoTenantContextError(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	tenant := model.Tenant{
		ID:     uuid.New().String(),
		Name:   "Tenant 1",
		Slug:   stringPtr("tenant1"),
		Status: "active",
	}
	db.Create(&tenant)

	user := model.User{
		ID:       uuid.New().String(),
		TenantID: &tenant.ID,
		Username: "user1",
		Email:    "user1@example.com",
		Fullname: "User 1",
	}
	db.Create(&user)

	authAdapter := postgresadapter.NewAuthAdapter(db)

	// Test: Query without tenant context should fail gracefully
	ctx := context.Background()
	foundUser, err := authAdapter.FindUserByEmail(ctx, "user1@example.com")
	assert.Error(t, err)
	assert.Nil(t, foundUser)
}

func TestTenantIsolation_CustomerNoContextError(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	tenant := model.Tenant{
		ID:     uuid.New().String(),
		Name:   "Tenant 1",
		Slug:   stringPtr("tenant1"),
		Status: "active",
	}
	db.Create(&tenant)

	customer := model.Customer{
		ID:              uuid.New().String(),
		TenantID:        tenant.ID,
		MikrotikID:      uuid.New().String(),
		ServiceUsername: "customer1",
		Name:            "Customer 1",
		Status:          model.CustomerStatusActive,
	}
	db.Create(&customer)

	customerAdapter := postgresadapter.NewCustomerAdapter(db)

	// Test: Customer lookup without tenant context should error
	ctx := context.Background()
	foundCustomer, err := customerAdapter.GetByPPPoEUsername(ctx, "customer1")
	assert.Error(t, err)
	assert.Nil(t, foundCustomer)
}
