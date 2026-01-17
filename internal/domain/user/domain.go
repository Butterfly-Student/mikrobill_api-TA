package user

import (
	"context"

	"MikrOps/internal/model"
	outbound_port "MikrOps/internal/port/outbound"

	"github.com/google/uuid"
	"github.com/palantir/stacktrace"
	"golang.org/x/crypto/bcrypt"
)

type UserDomain interface {
	CreateUser(ctx context.Context, input model.CreateUserRequest, createdBy string) (*model.User, error)
	GetUserByID(ctx context.Context, userID, requestingUserID string, isSuperAdmin bool) (*model.User, error)
	ListUsers(ctx context.Context, tenantID *string, requestingUserID string, isSuperAdmin bool, limit, offset int) ([]model.User, int64, error)
	UpdateUser(ctx context.Context, userID string, input model.UpdateUserRequest, updatedBy string, isSuperAdmin bool) (*model.User, error)
	DeleteUser(ctx context.Context, userID, requestingUserID string, isSuperAdmin bool) error
	AssignRole(ctx context.Context, userID, roleID, requestingUserID string, isSuperAdmin bool) error
	AssignToTenant(ctx context.Context, userID, tenantID string, roleID *string, isPrimary bool, requestingUserID string) error
}

type userDomain struct {
	databasePort outbound_port.DatabasePort
}

func NewUserDomain(databasePort outbound_port.DatabasePort) UserDomain {
	return &userDomain{
		databasePort: databasePort,
	}
}

// CreateUser creates a new user (super admin only)
func (d *userDomain) CreateUser(ctx context.Context, input model.CreateUserRequest, createdBy string) (*model.User, error) {
	db := d.databasePort.User()
	authDB := d.databasePort.Auth()

	// Check if email already exists
	existingUser, err := authDB.FindUserByEmail(ctx, input.Email)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to check existing email")
	}
	if existingUser != nil {
		return nil, stacktrace.NewError("email already exists")
	}

	// Check if username already exists
	existingUser, err = authDB.FindUserByUsername(ctx, input.Username)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to check existing username")
	}
	if existingUser != nil {
		return nil, stacktrace.NewError("username already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to hash password")
	}

	// Create user
	userRole := model.UserRoleViewer
	if input.UserRole != nil {
		userRole = *input.UserRole
	}

	status := model.UserStatusActive
	if input.Status != nil {
		status = *input.Status
	}

	user := &model.User{
		ID:                uuid.New().String(),
		Username:          input.Username,
		Email:             input.Email,
		EncryptedPassword: string(hashedPassword),
		Fullname:          input.Fullname,
		Phone:             input.Phone,
		UserRole:          userRole,
		Status:            status,
		RoleID:            input.RoleID,
		CreatedBy:         &createdBy,
	}

	if err := db.CreateUser(ctx, user); err != nil {
		return nil, stacktrace.Propagate(err, "failed to create user")
	}

	return user, nil
}

// GetUserByID retrieves a user by ID with access control
func (d *userDomain) GetUserByID(ctx context.Context, userID, requestingUserID string, isSuperAdmin bool) (*model.User, error) {
	db := d.databasePort.User()

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid user ID")
	}

	user, err := db.GetUserByID(ctx, userUUID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get user")
	}

	if user == nil {
		return nil, stacktrace.NewError("user not found")
	}

	// If not super admin, check if user has access to this user's tenant
	if !isSuperAdmin {
		if user.TenantID == nil {
			return nil, stacktrace.NewError("access denied")
		}

		requestingUUID, _ := uuid.Parse(requestingUserID)
		tenantUUID, _ := uuid.Parse(*user.TenantID)

		tenantUserDB := d.databasePort.TenantUser()
		hasAccess, err := tenantUserDB.HasAccess(ctx, requestingUUID, tenantUUID)
		if err != nil || !hasAccess {
			return nil, stacktrace.NewError("access denied")
		}
	}

	return user, nil
}

// ListUsers retrieves users with access control
func (d *userDomain) ListUsers(ctx context.Context, tenantID *string, requestingUserID string, isSuperAdmin bool, limit, offset int) ([]model.User, int64, error) {
	db := d.databasePort.User()

	var tenantUUID *uuid.UUID

	// If not super admin, force filter by requesting user's tenant
	if !isSuperAdmin {
		requestingUUID, _ := uuid.Parse(requestingUserID)
		tenantUserDB := d.databasePort.TenantUser()

		primaryTenant, err := tenantUserDB.GetPrimaryTenant(ctx, requestingUUID)
		if err != nil {
			return nil, 0, stacktrace.Propagate(err, "failed to get primary tenant")
		}
		tenantUUID = &primaryTenant
	} else if tenantID != nil {
		parsed, err := uuid.Parse(*tenantID)
		if err != nil {
			return nil, 0, stacktrace.Propagate(err, "invalid tenant ID")
		}
		tenantUUID = &parsed
	}

	users, total, err := db.ListUsers(ctx, tenantUUID, limit, offset)
	if err != nil {
		return nil, 0, stacktrace.Propagate(err, "failed to list users")
	}

	return users, total, nil
}

// UpdateUser updates a user with access control
func (d *userDomain) UpdateUser(ctx context.Context, userID string, input model.UpdateUserRequest, updatedBy string, isSuperAdmin bool) (*model.User, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid user ID")
	}

	// Get existing user
	existingUser, err := d.GetUserByID(ctx, userID, updatedBy, isSuperAdmin)
	if err != nil {
		return nil, err // Access control already checked
	}

	// Update fields
	if input.Fullname != nil {
		existingUser.Fullname = *input.Fullname
	}
	if input.Phone != nil {
		existingUser.Phone = input.Phone
	}
	if input.Avatar != nil {
		existingUser.Avatar = input.Avatar
	}
	if input.Status != nil {
		existingUser.Status = *input.Status
	}
	if input.UserRole != nil {
		existingUser.UserRole = *input.UserRole
	}
	if input.RoleID != nil {
		existingUser.RoleID = input.RoleID
	}

	existingUser.UpdatedBy = &updatedBy

	db := d.databasePort.User()
	if err := db.UpdateUser(ctx, existingUser); err != nil {
		return nil, stacktrace.Propagate(err, "failed to update user")
	}

	return db.GetUserByID(ctx, userUUID)
}

// DeleteUser soft deletes a user with access control
func (d *userDomain) DeleteUser(ctx context.Context, userID, requestingUserID string, isSuperAdmin bool) error {
	// Check access
	_, err := d.GetUserByID(ctx, userID, requestingUserID, isSuperAdmin)
	if err != nil {
		return err
	}

	userUUID, _ := uuid.Parse(userID)
	db := d.databasePort.User()

	if err := db.DeleteUser(ctx, userUUID); err != nil {
		return stacktrace.Propagate(err, "failed to delete user")
	}

	return nil
}

// AssignRole assigns a role to a user
func (d *userDomain) AssignRole(ctx context.Context, userID, roleID, requestingUserID string, isSuperAdmin bool) error {
	// Check access to user
	_, err := d.GetUserByID(ctx, userID, requestingUserID, isSuperAdmin)
	if err != nil {
		return err
	}

	userUUID, _ := uuid.Parse(userID)
	roleUUID, err := uuid.Parse(roleID)
	if err != nil {
		return stacktrace.Propagate(err, "invalid role ID")
	}

	db := d.databasePort.User()
	if err := db.AssignUserRole(ctx, userUUID, roleUUID); err != nil {
		return stacktrace.Propagate(err, "failed to assign role")
	}

	return nil
}

// AssignToTenant assigns a user to a tenant (super admin only)
func (d *userDomain) AssignToTenant(ctx context.Context, userID, tenantID string, roleID *string, isPrimary bool, requestingUserID string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return stacktrace.Propagate(err, "invalid user ID")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return stacktrace.Propagate(err, "invalid tenant ID")
	}

	var roleUUID *uuid.UUID
	if roleID != nil {
		parsed, err := uuid.Parse(*roleID)
		if err != nil {
			return stacktrace.Propagate(err, "invalid role ID")
		}
		roleUUID = &parsed
	}

	tenantUserDB := d.databasePort.TenantUser()
	if err := tenantUserDB.AssignUserToTenant(ctx, userUUID, tenantUUID, roleUUID, isPrimary); err != nil {
		return stacktrace.Propagate(err, "failed to assign user to tenant")
	}

	return nil
}
