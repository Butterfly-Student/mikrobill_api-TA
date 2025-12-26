package usecase

import (
	"context"
	"errors"
	"mikrobill/internal/entity"
	"mikrobill/internal/model"
	"mikrobill/internal/port/repository"
	"mikrobill/internal/port/service"
	pkg_logger "mikrobill/pkg/logger"
	"mikrobill/pkg/utils"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AuthUsecase defines the interface for authentication business logic
type AuthUsecase interface {
	Login(ctx context.Context, req model.LoginRequest) (*model.LoginResponse, error)
	Register(ctx context.Context, req model.CreateUserRequest) (*entity.User, error)
	ChangePassword(ctx context.Context, userID int64, req model.ChangePasswordRequest) error
	Logout(ctx context.Context, userID int64, token string) error
}

type authUsecase struct {
	userRepo        repository.UserRepository
	roleRepo        repository.RoleRepository
	passwordService *service.PasswordService
	jwtService      *service.JWTService
}

// NewAuthUsecase creates a new instance of AuthUsecase
func NewAuthUsecase(
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
	passwordService *service.PasswordService,
	jwtService *service.JWTService,
) AuthUsecase {
	return &authUsecase{
		userRepo:        userRepo,
		roleRepo:        roleRepo,
		passwordService: passwordService,
		jwtService:      jwtService,
	}
}

// Login handles user authentication and token generation
func (uc *authUsecase) Login(ctx context.Context, req model.LoginRequest) (*model.LoginResponse, error) {
	// Retrieve user by email
	user, err := uc.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrInvalidCredentials
		}
		return nil, err
	}

	pkg_logger.Debug("Login attempt",
		zap.String("email", req.Email),
		zap.Int("password_length", len(req.Password)),
	)

	// Verify password using password service
	if err := uc.passwordService.Verify(user.EncryptedPassword, req.Password); err != nil {
		// Increment failed login attempts
		_ = uc.userRepo.IncrementFailedLogin(ctx, user.ID)
		pkg_logger.Warn("Failed login attempt",
			zap.String("email", req.Email),
			zap.Int("attempts", user.FailedLoginAttempts+1),
		)
		return nil, utils.ErrInvalidCredentials
	}

	// Validate user account status
	if err := uc.validateUserAccount(user); err != nil {
		return nil, err
	}

	// Resolve user role
	roleName, err := uc.resolveUserRole(ctx, user)
	if err != nil {
		pkg_logger.Warn("Failed to resolve user role", zap.Error(err))
		roleName = string(user.UserRole)
	}

	// Reset failed login attempts
	if err := uc.userRepo.ResetFailedLogin(ctx, user.ID); err != nil {
		pkg_logger.Warn("Failed to reset login attempts", zap.Error(err))
	}

	// Update last login
	clientIP := req.IP
	if clientIP == "" {
		clientIP = "unknown"
	}
	if err := uc.userRepo.UpdateLastLogin(ctx, user.ID, clientIP); err != nil {
		pkg_logger.Warn("Failed to update last login", zap.Error(err))
	}

	// Generate JWT token using JWT service
	token, expiresAt, err := uc.jwtService.GenerateToken(user.ID, user.Email, roleName)
	if err != nil {
		return nil, err
	}

	pkg_logger.Info("User logged in successfully",
		zap.Int64("user_id", user.ID),
		zap.String("email", user.Email),
		zap.String("role", roleName),
	)

	return &model.LoginResponse{
		Token:     token,
		ExpiresAt: time.Unix(expiresAt, 0),
		User: model.UserSummary{
			ID:     user.ID,
			Name:   user.Fullname,
			Email:  user.Email,
			Status: string(user.Status),
			Roles:  []string{roleName},
		},
	}, nil
}

// Register handles user registration
func (uc *authUsecase) Register(ctx context.Context, req model.CreateUserRequest) (*entity.User, error) {
	// Validate email uniqueness
	if err := uc.validateEmailUniqueness(ctx, req.Email); err != nil {
		return nil, err
	}

	// Validate username uniqueness if provided
	if req.Username != "" {
		if err := uc.validateUsernameUniqueness(ctx, req.Username); err != nil {
			return nil, err
		}
	}

	// Hash password using password service
	hashedPassword, err := uc.passwordService.Hash(req.Password)
	if err != nil {
		return nil, err
	}

	// Prepare user entity
	user, err := uc.buildUserEntity(ctx, req, hashedPassword)
	if err != nil {
		return nil, err
	}

	// Persist user to database
	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	pkg_logger.Info("User registered successfully",
		zap.Int64("user_id", user.ID),
		zap.String("email", user.Email),
		zap.String("user_role", string(user.UserRole)),
	)

	return user, nil
}

// Logout handles user logout and token invalidation
func (uc *authUsecase) Logout(ctx context.Context, userID int64, token string) error {
	// Log the logout activity
	pkg_logger.Info("User logged out",
		zap.Int64("user_id", userID),
	)

	// TODO: Implement token blacklist/revocation if needed
	// Options:
	// 1. Store revoked tokens in Redis with TTL
	// 2. Store revoked tokens in database
	// 3. Use token versioning in user table
	// 4. Simply log the logout (current implementation)

	// For now, we just log the logout
	// JWT tokens will naturally expire based on their expiration time
	// If you need immediate invalidation, implement one of the above options

	return nil
}

// ChangePassword handles password change for authenticated user
func (uc *authUsecase) ChangePassword(ctx context.Context, userID int64, req model.ChangePasswordRequest) error {
	// Retrieve user by ID
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrUserNotFound
		}
		return err
	}

	// Verify old password using password service
	if err := uc.passwordService.Verify(user.EncryptedPassword, req.OldPassword); err != nil {
		return errors.New("old password is incorrect")
	}

	// Hash new password using password service
	hashedPassword, err := uc.passwordService.Hash(req.NewPassword)
	if err != nil {
		return err
	}

	// Update password in repository
	if err := uc.userRepo.UpdatePassword(ctx, userID, hashedPassword); err != nil {
		return err
	}

	pkg_logger.Info("Password changed successfully",
		zap.Int64("user_id", userID),
	)

	return nil
}

// validateUserAccount checks if user account is active and not locked
func (uc *authUsecase) validateUserAccount(user *entity.User) error {
	if user.Status != entity.UserStatusActive {
		return errors.New("user account is not active")
	}

	if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
		return errors.New("account is temporarily locked")
	}

	return nil
}

// resolveUserRole determines the role name for the user
func (uc *authUsecase) resolveUserRole(ctx context.Context, user *entity.User) (string, error) {
	// Priority 1: Use role from role_id if available
	if user.RoleID != nil {
		role, err := uc.roleRepo.GetByID(ctx, *user.RoleID)
		if err == nil {
			return role.Name, nil
		}
		pkg_logger.Warn("Failed to get role by ID",
			zap.Error(err),
			zap.Int64("role_id", *user.RoleID),
		)
	}

	// Priority 2: Fall back to UserRole enum
	return string(user.UserRole), nil
}

// validateEmailUniqueness checks if email is already registered
func (uc *authUsecase) validateEmailUniqueness(ctx context.Context, email string) error {
	exists, err := uc.userRepo.ExistsByEmail(ctx, email)
	if err != nil {
		return err
	}
	if exists {
		return utils.ErrUserAlreadyExists
	}
	return nil
}

// validateUsernameUniqueness checks if username is already taken
func (uc *authUsecase) validateUsernameUniqueness(ctx context.Context, username string) error {
	exists, err := uc.userRepo.ExistsByUsername(ctx, username)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("username already exists")
	}
	return nil
}

// buildUserEntity constructs a user entity from registration request
func (uc *authUsecase) buildUserEntity(ctx context.Context, req model.CreateUserRequest, hashedPassword string) (*entity.User, error) {
	// Set default status
	status := entity.UserStatusActive
	if req.Status != "" {
		status = entity.UserStatus(req.Status)
	}

	// Set default user role
	userRole := entity.UserRoleViewer
	if req.UserRole != "" {
		userRole = entity.UserRole(req.UserRole)
	}

	// Prepare user entity
	user := &entity.User{
		Username:          req.Username,
		Email:             req.Email,
		EncryptedPassword: hashedPassword,
		Fullname:          req.Name,
		Phone:             req.Phone,
		UserRole:          userRole,
		Status:            status,
	}

	// Validate and assign role if provided
	if len(req.RoleIDs) > 0 {
		role, err := uc.roleRepo.GetByID(ctx, req.RoleIDs[0])
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("role not found")
			}
			return nil, err
		}
		user.RoleID = &role.ID
	}

	return user, nil
}