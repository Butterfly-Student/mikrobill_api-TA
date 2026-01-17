package auth

import (
	"context"
	"os"
	"time"

	"MikrOps/internal/model"
	outbound_port "MikrOps/internal/port/outbound"
	"MikrOps/utils"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/palantir/stacktrace"
	"golang.org/x/crypto/bcrypt"
)

// Token configuration constants
const (
	AccessTokenTTL     = 15 * time.Minute    // Short-lived access token
	RefreshTokenTTL    = 30 * 24 * time.Hour // 30 days
	AbsoluteMaxSession = 90 * 24 * time.Hour // 90 days hard limit
	RotationThreshold  = 3 * 24 * time.Hour  // Rotate if < 3 days left
)

type AuthDomain interface {
	Register(ctx context.Context, input model.RegisterRequest) (*model.User, error)
	Login(ctx context.Context, identifier, password string) (*model.User, string, string, error)
	Logout(ctx context.Context, userID string, refreshToken string) error
	RefreshToken(ctx context.Context, refreshToken string) (*model.User, string, string, error)
	GetUserProfile(ctx context.Context, userID string) (*model.User, error)
	ValidateToken(ctx context.Context, token string) (*model.User, error)
}

type authDomain struct {
	databasePort outbound_port.DatabasePort
	cachePort    outbound_port.CachePort
}

func NewAuthDomain(databasePort outbound_port.DatabasePort, cachePort outbound_port.CachePort) AuthDomain {
	return &authDomain{
		databasePort: databasePort,
		cachePort:    cachePort,
	}
}

func (s *authDomain) Register(ctx context.Context, input model.RegisterRequest) (*model.User, error) {
	// Check if user exists
	db := s.databasePort.Auth()

	existingUser, err := db.FindUserByEmail(ctx, input.Email)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to check existing email")
	}
	if existingUser != nil {
		return nil, stacktrace.NewError("email already registered")
	}

	existingUser, err = db.FindUserByUsername(ctx, input.Username)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to check existing username")
	}
	if existingUser != nil {
		return nil, stacktrace.NewError("username already taken")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to hash password")
	}

	roleIDStr := ""
	userRole := model.UserRoleViewer // Default

	// Find default role
	role, err := db.FindRoleByName(ctx, string(model.UserRoleViewer))
	if err == nil && role != nil {
		roleIDStr = role.ID
		userRole = model.UserRoleViewer
	}

	newUser := model.User{
		ID:                uuid.New().String(),
		Username:          input.Username,
		Email:             input.Email,
		EncryptedPassword: string(hashed),
		Fullname:          input.Fullname,
		Status:            model.UserStatusActive,
		RoleID:            &roleIDStr,
		UserRole:          userRole,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if input.Phone != "" {
		newUser.Phone = &input.Phone
	}

	if err := db.SaveUser(ctx, newUser); err != nil {
		return nil, stacktrace.Propagate(err, "failed to save user")
	}

	return &newUser, nil
}

func (s *authDomain) Login(ctx context.Context, identifier, password string) (*model.User, string, string, error) {
	db := s.databasePort.Auth()

	// Find user by email OR username in single query
	user, err := db.FindUserByEmailOrUsername(ctx, identifier)
	if err != nil {
		return nil, "", "", stacktrace.NewError("invalid credentials")
	}

	if user == nil {
		return nil, "", "", stacktrace.NewError("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.EncryptedPassword), []byte(password)); err != nil {
		return nil, "", "", stacktrace.NewError("invalid credentials")
	}

	if user.Status != model.UserStatusActive {
		return nil, "", "", stacktrace.NewError("account is not active")
	}

	// Generate token pair (access + refresh)
	accessToken, refreshToken, err := s.generateTokenPair(ctx, user)
	if err != nil {
		return nil, "", "", stacktrace.Propagate(err, "failed to generate tokens")
	}

	return user, accessToken, refreshToken, nil
}

// Logout removes refresh token from Redis
func (s *authDomain) Logout(ctx context.Context, userID string, refreshToken string) error {
	cache := s.cachePort.AuthCache()

	// Hash the token
	tokenHash := utils.HashToken(refreshToken)

	// Invalidate the specific token
	if err := cache.InvalidateToken(ctx, tokenHash); err != nil {
		return stacktrace.Propagate(err, "failed to invalidate refresh token")
	}

	return nil
}

// RefreshToken validates refresh token and generates new tokens with rotation
func (s *authDomain) RefreshToken(ctx context.Context, refreshToken string) (*model.User, string, string, error) {
	cache := s.cachePort.AuthCache()

	// Hash the incoming token
	tokenHash := utils.HashToken(refreshToken)

	// REUSE DETECTION: Check if token was already rotated
	isRotated, err := cache.IsTokenRotated(ctx, tokenHash)
	if err != nil {
		return nil, "", "", stacktrace.Propagate(err, "failed to check token rotation status")
	}

	if isRotated {
		// SECURITY BREACH: Token reuse detected!
		metadata, _ := cache.GetRefreshTokenMetadata(ctx, tokenHash)
		if metadata != nil {
			// Invalidate ALL user tokens as security measure
			_ = cache.InvalidateAllUserTokens(ctx, metadata.UserID)
		}
		return nil, "", "", stacktrace.NewError("token reuse detected - all sessions invalidated")
	}

	// Get token metadata
	metadata, err := cache.GetRefreshTokenMetadata(ctx, tokenHash)
	if err != nil {
		return nil, "", "", stacktrace.NewError("invalid or expired refresh token")
	}

	// Validate absolute expiry
	if time.Now().After(metadata.AbsoluteExpiry) {
		_ = cache.InvalidateToken(ctx, tokenHash)
		return nil, "", "", stacktrace.NewError("absolute session limit exceeded - please login again")
	}

	// Get user from database
	userUUID, err := uuid.Parse(metadata.UserID)
	if err != nil {
		return nil, "", "", stacktrace.Propagate(err, "invalid user id")
	}

	db := s.databasePort.Auth()
	user, err := db.FindUserByID(ctx, userUUID)
	if err != nil || user == nil {
		return nil, "", "", stacktrace.NewError("user not found")
	}

	if user.Status != model.UserStatusActive {
		return nil, "", "", stacktrace.NewError("account is not active")
	}

	// Generate new access token
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, "", "", stacktrace.Propagate(err, "failed to generate access token")
	}

	// ROTATION DECISION: Check if token is approaching expiry
	timeUntilExpiry := time.Until(metadata.ExpiresAt)
	newRefreshToken := ""

	if timeUntilExpiry < RotationThreshold {
		// TOKEN ROTATION: Generate new refresh token

		// Mark old token as rotated (for reuse detection)
		if err := cache.MarkTokenRotated(ctx, tokenHash); err != nil {
			return nil, "", "", stacktrace.Propagate(err, "failed to mark token as rotated")
		}

		// Invalidate old token
		if err := cache.InvalidateToken(ctx, tokenHash); err != nil {
			return nil, "", "", stacktrace.Propagate(err, "failed to invalidate old token")
		}

		// Generate new refresh token with updated metadata
		newRefreshToken, err = s.generateRefreshTokenWithMetadata(ctx, user, &tokenHash, metadata.AbsoluteExpiry, metadata.RotationCount+1)
		if err != nil {
			return nil, "", "", stacktrace.Propagate(err, "failed to generate new refresh token")
		}
	} else {
		// NO ROTATION: Just update last_used
		if err := cache.MarkTokenUsed(ctx, tokenHash); err != nil {
			return nil, "", "", stacktrace.Propagate(err, "failed to update token usage")
		}
	}

	return user, accessToken, newRefreshToken, nil
}

// GetUserProfile retrieves user profile by ID
func (s *authDomain) GetUserProfile(ctx context.Context, userID string) (*model.User, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid user id")
	}

	db := s.databasePort.Auth()
	user, err := db.FindUserByID(ctx, userUUID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to find user")
	}

	if user == nil {
		return nil, stacktrace.NewError("user not found")
	}

	return user, nil
}

// generateAccessToken generates JWT access token with configurable TTL
func (s *authDomain) generateAccessToken(user *model.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":      user.ID,
		"username": user.Username,
		"role":     user.UserRole,
		"exp":      time.Now().Add(AccessTokenTTL).Unix(),
	})

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "default_secret_please_change"
	}

	return token.SignedString([]byte(secret))
}

// generateRefreshTokenWithMetadata generates refresh token with full metadata for rotation
func (s *authDomain) generateRefreshTokenWithMetadata(ctx context.Context, user *model.User, parentHash *string, absoluteExpiry time.Time, rotationCount int) (string, error) {
	// Generate a unique refresh token (UUID)
	refreshToken := uuid.New().String()

	// Hash it for storage
	tokenHash := utils.HashToken(refreshToken)

	now := time.Now()
	expiresAt := now.Add(RefreshTokenTTL)

	// Ensure absolute expiry doesn't extend beyond max
	if absoluteExpiry.IsZero() {
		// First time - set absolute expiry
		absoluteExpiry = now.Add(AbsoluteMaxSession)
	} else {
		// Rotation - keep same absolute expiry, don't extend
		if expiresAt.After(absoluteExpiry) {
			expiresAt = absoluteExpiry
		}
	}

	// Create metadata
	metadata := model.RefreshTokenMetadata{
		UserID:          user.ID,
		TokenHash:       tokenHash,
		IssuedAt:        now,
		ExpiresAt:       expiresAt,
		LastUsedAt:      now,
		RotationCount:   rotationCount,
		AbsoluteExpiry:  absoluteExpiry,
		ParentTokenHash: parentHash,
	}

	// Store in Redis
	cache := s.cachePort.AuthCache()
	if err := cache.StoreRefreshToken(ctx, metadata); err != nil {
		return "", stacktrace.Propagate(err, "failed to store refresh token")
	}

	return refreshToken, nil
}

// generateTokenPair generates both access and refresh tokens for initial login
func (s *authDomain) generateTokenPair(ctx context.Context, user *model.User) (accessToken, refreshToken string, err error) {
	accessToken, err = s.generateAccessToken(user)
	if err != nil {
		return "", "", err
	}

	// Generate refresh token with metadata (no parent, rotation count = 0)
	refreshToken, err = s.generateRefreshTokenWithMetadata(ctx, user, nil, time.Time{}, 0)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}
func (s *authDomain) ValidateToken(ctx context.Context, tokenString string) (*model.User, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "default_secret_please_change"
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, stacktrace.NewError("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to parse token")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userIDStr, ok := claims["sub"].(string)
		if !ok {
			return nil, stacktrace.NewError("invalid token claims")
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return nil, stacktrace.Propagate(err, "invalid user id in token")
		}

		// Check DB if user still exists/active
		db := s.databasePort.Auth()
		user, err := db.FindUserByID(ctx, userID)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to find user")
		}
		if user == nil {
			return nil, stacktrace.NewError("user not found")
		}

		if user.Status != model.UserStatusActive {
			return nil, stacktrace.NewError("user is inactive")
		}

		return user, nil
	}

	return nil, stacktrace.NewError("invalid token")
}
