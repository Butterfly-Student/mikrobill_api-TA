package customer_auth

import (
	"context"
	"os"
	"time"

	"MikrOps/internal/model"
	outbound_port "MikrOps/internal/port/outbound"
	"MikrOps/utils"
	"MikrOps/utils/encryption"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/palantir/stacktrace"
	"golang.org/x/crypto/bcrypt"
)

// Token configuration constants for customer portal
const (
	CustomerAccessTokenTTL     = 15 * time.Minute    // Short-lived access token
	CustomerRefreshTokenTTL    = 30 * 24 * time.Hour // 30 days
	CustomerAbsoluteMaxSession = 90 * 24 * time.Hour // 90 days hard limit
	CustomerRotationThreshold  = 3 * 24 * time.Hour  // Rotate if < 3 days left
)

// CustomerAuthDomain provides customer authentication services for the customer portal
type CustomerAuthDomain interface {
	CustomerLogin(ctx context.Context, tenantID uuid.UUID, email, password string) (*model.Customer, string, string, error)
	CustomerLogout(ctx context.Context, customerID string, refreshToken string) error
	RefreshToken(ctx context.Context, refreshToken string) (*model.Customer, string, string, error)
	GetCustomerProfile(ctx context.Context, customerID uuid.UUID) (*model.CustomerPortalProfileResponse, error)
	ValidateToken(ctx context.Context, token string) (*model.Customer, error)
}

type customerAuthDomain struct {
	customerPort  outbound_port.CustomerDatabasePort
	cachePort     outbound_port.CachePort
	encryptionSvc *encryption.Service
}

func NewCustomerAuthDomain(
	customerPort outbound_port.CustomerDatabasePort,
	cachePort outbound_port.CachePort,
	encryptionSvc *encryption.Service,
) CustomerAuthDomain {
	return &customerAuthDomain{
		customerPort:  customerPort,
		cachePort:     cachePort,
		encryptionSvc: encryptionSvc,
	}
}

// CustomerLogin authenticates customer using portal credentials (email + password)
func (d *customerAuthDomain) CustomerLogin(ctx context.Context, tenantID uuid.UUID, email, password string) (*model.Customer, string, string, error) {
	// Find customer by portal_email
	customer, err := d.customerPort.GetByPortalEmail(ctx, tenantID, email)
	if err != nil {
		return nil, "", "", stacktrace.NewError("invalid credentials")
	}

	if customer == nil {
		return nil, "", "", stacktrace.NewError("invalid credentials")
	}

	// Verify portal password
	if customer.PortalPasswordHash == nil {
		return nil, "", "", stacktrace.NewError("customer has no portal password set")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*customer.PortalPasswordHash), []byte(password)); err != nil {
		return nil, "", "", stacktrace.NewError("invalid credentials")
	}

	// Check if customer is active
	if customer.Status != model.CustomerStatusActive {
		return nil, "", "", stacktrace.NewError("account is not active")
	}

	// Generate token pair (access + refresh)
	accessToken, refreshToken, err := d.generateTokenPair(ctx, customer)
	if err != nil {
		return nil, "", "", stacktrace.Propagate(err, "failed to generate tokens")
	}

	return customer, accessToken, refreshToken, nil
}

// CustomerLogout removes refresh token from Redis
func (d *customerAuthDomain) CustomerLogout(ctx context.Context, customerID string, refreshToken string) error {
	cache := d.cachePort.AuthCache()

	// Hash the token
	tokenHash := utils.HashToken(refreshToken)

	// Invalidate the specific token
	if err := cache.InvalidateToken(ctx, tokenHash); err != nil {
		return stacktrace.Propagate(err, "failed to invalidate refresh token")
	}

	return nil
}

// RefreshToken validates refresh token and generates new tokens with rotation
func (d *customerAuthDomain) RefreshToken(ctx context.Context, refreshToken string) (*model.Customer, string, string, error) {
	cache := d.cachePort.AuthCache()

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
			// Invalidate ALL customer tokens as security measure
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

	// Get customer from database
	customerUUID, err := uuid.Parse(metadata.UserID)
	if err != nil {
		return nil, "", "", stacktrace.Propagate(err, "invalid customer id")
	}

	customer, err := d.customerPort.GetByID(ctx, customerUUID)
	if err != nil || customer == nil {
		return nil, "", "", stacktrace.NewError("customer not found")
	}

	if customer.Status != model.CustomerStatusActive {
		return nil, "", "", stacktrace.NewError("account is not active")
	}

	// Generate new access token
	accessToken, err := d.generateAccessToken(customer)
	if err != nil {
		return nil, "", "", stacktrace.Propagate(err, "failed to generate access token")
	}

	// ROTATION DECISION: Check if token is approaching expiry
	timeUntilExpiry := time.Until(metadata.ExpiresAt)
	newRefreshToken := ""

	if timeUntilExpiry < CustomerRotationThreshold {
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
		newRefreshToken, err = d.generateRefreshTokenWithMetadata(ctx, customer, &tokenHash, metadata.AbsoluteExpiry, metadata.RotationCount+1)
		if err != nil {
			return nil, "", "", stacktrace.Propagate(err, "failed to generate new refresh token")
		}
	} else {
		// NO ROTATION: Just update last_used
		if err := cache.MarkTokenUsed(ctx, tokenHash); err != nil {
			return nil, "", "", stacktrace.Propagate(err, "failed to update token usage")
		}
	}

	return customer, accessToken, newRefreshToken, nil
}

// GetCustomerProfile retrieves customer profile with service credentials (if visible)
func (d *customerAuthDomain) GetCustomerProfile(ctx context.Context, customerID uuid.UUID) (*model.CustomerPortalProfileResponse, error) {
	customer, err := d.customerPort.GetByID(ctx, customerID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get customer")
	}

	if customer == nil {
		return nil, stacktrace.NewError("customer not found")
	}

	// Build response
	response := &model.CustomerPortalProfileResponse{
		ID:              customer.ID,
		Name:            customer.Name,
		Phone:           customer.Phone,
		Email:           customer.Email,
		Address:         customer.Address,
		ServiceType:     customer.ServiceType,
		Status:          string(customer.Status),
		JoinDate:        customer.JoinDate.Format("2006-01-02"),
		ServiceUsername: customer.ServiceUsername,
		AssignedIP:      customer.AssignedIP,
		LastOnline:      customer.LastOnline,
	}

	// Decrypt and show service password ONLY if visible (Hotspot)
	if customer.ServicePasswordVisible && customer.ServicePasswordEncrypted != nil {
		decryptedPassword, err := d.encryptionSvc.Decrypt(*customer.ServicePasswordEncrypted)
		if err == nil && decryptedPassword != "" {
			response.ServicePassword = &decryptedPassword
		}
	}

	// Add current service if available
	if len(customer.Services) > 0 {
		currentService := customer.Services[0].ToResponse()
		response.CurrentPlan = currentService
	}

	return response, nil
}

// ValidateToken validates customer JWT access token
func (d *customerAuthDomain) ValidateToken(ctx context.Context, tokenString string) (*model.Customer, error) {
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
		customerIDStr, ok := claims["sub"].(string)
		if !ok {
			return nil, stacktrace.NewError("invalid token claims")
		}

		// Verify this is a customer token (has customer_type claim)
		tokenType, _ := claims["type"].(string)
		if tokenType != "customer" {
			return nil, stacktrace.NewError("invalid token type")
		}

		customerID, err := uuid.Parse(customerIDStr)
		if err != nil {
			return nil, stacktrace.Propagate(err, "invalid customer id in token")
		}

		// Check DB if customer still exists/active
		customer, err := d.customerPort.GetByID(ctx, customerID)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to find customer")
		}
		if customer == nil {
			return nil, stacktrace.NewError("customer not found")
		}

		if customer.Status != model.CustomerStatusActive {
			return nil, stacktrace.NewError("customer is inactive")
		}

		return customer, nil
	}

	return nil, stacktrace.NewError("invalid token")
}

// generateAccessToken generates JWT access token for customer
func (d *customerAuthDomain) generateAccessToken(customer *model.Customer) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":       customer.ID,
		"type":      "customer", // Differentiate from admin tokens
		"tenant_id": customer.TenantID,
		"email":     customer.PortalEmail,
		"exp":       time.Now().Add(CustomerAccessTokenTTL).Unix(),
	})

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "default_secret_please_change"
	}

	return token.SignedString([]byte(secret))
}

// generateRefreshTokenWithMetadata generates refresh token with full metadata for rotation
func (d *customerAuthDomain) generateRefreshTokenWithMetadata(ctx context.Context, customer *model.Customer, parentHash *string, absoluteExpiry time.Time, rotationCount int) (string, error) {
	// Generate a unique refresh token (UUID)
	refreshToken := uuid.New().String()

	// Hash it for storage
	tokenHash := utils.HashToken(refreshToken)

	now := time.Now()
	expiresAt := now.Add(CustomerRefreshTokenTTL)

	// Ensure absolute expiry doesn't extend beyond max
	if absoluteExpiry.IsZero() {
		// First time - set absolute expiry
		absoluteExpiry = now.Add(CustomerAbsoluteMaxSession)
	} else {
		// Rotation - keep same absolute expiry, don't extend
		if expiresAt.After(absoluteExpiry) {
			expiresAt = absoluteExpiry
		}
	}

	// Create metadata (use customer ID as UserID for compatibility with auth cache)
	metadata := model.RefreshTokenMetadata{
		UserID:          customer.ID,
		TokenHash:       tokenHash,
		IssuedAt:        now,
		ExpiresAt:       expiresAt,
		LastUsedAt:      now,
		RotationCount:   rotationCount,
		AbsoluteExpiry:  absoluteExpiry,
		ParentTokenHash: parentHash,
	}

	// Store in Redis
	cache := d.cachePort.AuthCache()
	if err := cache.StoreRefreshToken(ctx, metadata); err != nil {
		return "", stacktrace.Propagate(err, "failed to store refresh token")
	}

	return refreshToken, nil
}

// generateTokenPair generates both access and refresh tokens for initial login
func (d *customerAuthDomain) generateTokenPair(ctx context.Context, customer *model.Customer) (accessToken, refreshToken string, err error) {
	accessToken, err = d.generateAccessToken(customer)
	if err != nil {
		return "", "", err
	}

	// Generate refresh token with metadata (no parent, rotation count = 0)
	refreshToken, err = d.generateRefreshTokenWithMetadata(ctx, customer, nil, time.Time{}, 0)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}
