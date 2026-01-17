package auth

import (
	"context"
	"os"
	"time"

	"MikrOps/internal/model"
	outbound_port "MikrOps/internal/port/outbound"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/palantir/stacktrace"
	"golang.org/x/crypto/bcrypt"
)

type AuthDomain interface {
	Register(ctx context.Context, input model.RegisterRequest) (*model.User, error)
	Login(ctx context.Context, identifier, password string) (*model.User, string, error)
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

func (s *authDomain) Login(ctx context.Context, identifier, password string) (*model.User, string, error) {
	db := s.databasePort.Auth()

	// Try to find by email first
	user, err := db.FindUserByEmail(ctx, identifier)
	if err != nil {
		return nil, "", stacktrace.Propagate(err, "failed to find user by email")
	}

	// If not found by email, try username
	if user == nil {
		user, err = db.FindUserByUsername(ctx, identifier)
		if err != nil {
			return nil, "", stacktrace.Propagate(err, "failed to find user by username")
		}
	}

	if user == nil {
		return nil, "", stacktrace.NewError("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.EncryptedPassword), []byte(password)); err != nil {
		return nil, "", stacktrace.NewError("invalid credentials")
	}

	if user.Status != model.UserStatusActive {
		return nil, "", stacktrace.NewError("account is not active")
	}

	// Generate JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":      user.ID,
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
		return nil, "", stacktrace.Propagate(err, "failed to sign token")
	}

	return user, tokenString, nil
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

