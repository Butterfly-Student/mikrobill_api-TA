package auth

import (
	"context"
	"os"
	"time"

	"prabogo/internal/model"
	outbound_port "prabogo/internal/port/outbound"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/palantir/stacktrace"
	"golang.org/x/crypto/bcrypt"
)

type AuthDomain interface {
	Register(ctx context.Context, input model.UserInput) (*model.User, error)
	Login(ctx context.Context, email, password string) (string, error)
	ValidateToken(ctx context.Context, token string) (*model.User, error)
}

type authDomain struct {
	databasePort outbound_port.DatabasePort
}

func NewAuthDomain(databasePort outbound_port.DatabasePort) AuthDomain {
	return &authDomain{
		databasePort: databasePort,
	}
}

func (s *authDomain) Register(ctx context.Context, input model.UserInput) (*model.User, error) {
	// Check if user exists
	db := s.databasePort.Auth()
	users, err := db.FindUserByFilter(model.UserFilter{Emails: []string{input.Email}}, false)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to check existing user")
	}
	if len(users) > 0 {
		return nil, stacktrace.NewError("email already registered")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to hash password")
	}

	roleID := uuid.Nil
	userRole := model.UserRoleViewer // Default

	// If input.RoleID is provided, validate it?
	// For now, let's look up role by name if needed or assume user provides valid RoleID string if generic
	// But basic registration usually defaults to 'viewer' or 'user' unless admin creates it.
	// For simplicity, I'll default to 'viewer' and find that role.

	// Find default role
	// This assumes roles are seeded.
	role, err := db.FindRoleByName("viewer")
	if err == nil && role != nil {
		roleID = role.ID
		userRole = model.UserRoleViewer // redundant but consistent
	}

	newUser := model.User{
		ID:                uuid.New(),
		Username:          input.Username,
		Email:             input.Email,
		EncryptedPassword: string(hashed),
		Fullname:          input.Fullname,
		Status:            model.UserStatusActive,
		RoleID:            &roleID,
		UserRole:          userRole,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if input.Phone != "" {
		newUser.Phone = &input.Phone
	}

	if err := db.SaveUser(newUser); err != nil {
		return nil, stacktrace.Propagate(err, "failed to save user")
	}

	return &newUser, nil
}

func (s *authDomain) Login(ctx context.Context, email, password string) (string, error) {
	db := s.databasePort.Auth()
	users, err := db.FindUserByFilter(model.UserFilter{Emails: []string{email}}, false)
	if err != nil {
		return "", stacktrace.Propagate(err, "failed to find user")
	}
	if len(users) == 0 {
		return "", stacktrace.NewError("invalid credentials")
	}
	user := users[0]

	if err := bcrypt.CompareHashAndPassword([]byte(user.EncryptedPassword), []byte(password)); err != nil {
		return "", stacktrace.NewError("invalid credentials")
	}

	if user.Status != model.UserStatusActive {
		return "", stacktrace.NewError("account is not active")
	}

	// Generate JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":      user.ID.String(),
		"username": user.Username,
		"role":     user.UserRole,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "default_secret_please_change"
	}

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", stacktrace.Propagate(err, "failed to sign token")
	}

	return tokenString, nil
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

		// Optionally check DB if user still exists/active
		db := s.databasePort.Auth()
		users, err := db.FindUserByFilter(model.UserFilter{IDs: []uuid.UUID{userID}}, false)
		if err != nil || len(users) == 0 {
			return nil, stacktrace.NewError("user not found")
		}

		if users[0].Status != model.UserStatusActive {
			return nil, stacktrace.NewError("user is inactive")
		}

		return &users[0], nil
	}

	return nil, stacktrace.NewError("invalid token")
}
