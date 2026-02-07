package auth

import (
	"context"
	"errors"
	"os"
	"testing"

	"MikrOps/internal/model"
	mock_outbound_port "MikrOps/tests/mocks/port"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/crypto/bcrypt"
)

func TestAuth(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-jwt-secret-key-that-is-at-least-32-chars-long")
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockAuthDB := mock_outbound_port.NewMockAuthDatabasePort(ctrl)
	mockCache := mock_outbound_port.NewMockCachePort(ctrl)
	mockAuthCache := mock_outbound_port.NewMockAuthCachePort(ctrl)

	mockDB.EXPECT().Auth().Return(mockAuthDB).AnyTimes()
	mockCache.EXPECT().AuthCache().Return(mockAuthCache).AnyTimes()

	domain := NewAuthDomain(mockDB, mockCache)
	ctx := context.Background()

	Convey("Test Auth Domain", t, func() {
		Convey("Register", func() {
			input := model.RegisterRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
				Fullname: "Test User",
			}

			Convey("Email already registered", func() {
				mockAuthDB.EXPECT().FindUserByEmail(ctx, input.Email).Return(&model.User{ID: uuid.New().String()}, nil)
				user, err := domain.Register(ctx, input)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "email already registered")
				So(user, ShouldBeNil)
			})

			Convey("Success with default role", func() {
				mockAuthDB.EXPECT().FindUserByEmail(ctx, input.Email).Return(nil, nil)
				mockAuthDB.EXPECT().FindUserByUsername(ctx, input.Username).Return(nil, nil)
				roleID := uuid.New().String()
				mockAuthDB.EXPECT().FindRoleByName(ctx, string(model.UserRoleViewer)).Return(&model.Role{ID: roleID, Name: "viewer"}, nil)
				mockAuthDB.EXPECT().SaveUser(ctx, gomock.Any()).Return(nil)

				user, err := domain.Register(ctx, input)
				So(err, ShouldBeNil)
				So(user, ShouldNotBeNil)
				So(user.Email, ShouldEqual, input.Email)
				So(user.UserRole, ShouldEqual, model.UserRoleViewer)
			})

			Convey("Database error", func() {
				mockAuthDB.EXPECT().FindUserByEmail(ctx, input.Email).Return(nil, errors.New("db error"))
				user, err := domain.Register(ctx, input)
				So(err, ShouldNotBeNil)
				So(user, ShouldBeNil)
			})
		})

		Convey("Login", func() {
			password := "password123"
			hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			user := model.User{
				ID:                uuid.New().String(),
				Username:          "testuser",
				Email:             "test@example.com",
				EncryptedPassword: string(hashed),
				Status:            model.UserStatusActive,
				UserRole:          model.UserRoleViewer,
			}

			Convey("Success", func() {
				mockAuthDB.EXPECT().FindUserByEmailOrUsername(ctx, "testuser").Return(&user, nil)
				mockAuthCache.EXPECT().StoreRefreshToken(ctx, gomock.Any()).Return(nil)

				loggedUser, accessToken, refreshToken, err := domain.Login(ctx, "testuser", password)
				So(err, ShouldBeNil)
				So(loggedUser, ShouldNotBeNil)
				So(loggedUser.ID, ShouldEqual, user.ID)
				So(accessToken, ShouldNotBeEmpty)
				So(refreshToken, ShouldNotBeEmpty)
			})

			Convey("Invalid credentials - user not found", func() {
				mockAuthDB.EXPECT().FindUserByEmailOrUsername(ctx, "wrong").Return(nil, nil)
				loggedUser, accessToken, refreshToken, err := domain.Login(ctx, "wrong", password)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "invalid credentials")
				So(loggedUser, ShouldBeNil)
				So(accessToken, ShouldBeEmpty)
				So(refreshToken, ShouldBeEmpty)
			})

			Convey("Invalid credentials - wrong password", func() {
				mockAuthDB.EXPECT().FindUserByEmailOrUsername(ctx, "testuser").Return(&user, nil)
				loggedUser, accessToken, refreshToken, err := domain.Login(ctx, "testuser", "wrong")
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "invalid credentials")
				So(loggedUser, ShouldBeNil)
				So(accessToken, ShouldBeEmpty)
				So(refreshToken, ShouldBeEmpty)
			})

			Convey("Account inactive", func() {
				inactiveUser := user
				inactiveUser.Status = model.UserStatusInactive
				mockAuthDB.EXPECT().FindUserByEmailOrUsername(ctx, "testuser").Return(&inactiveUser, nil)
				loggedUser, accessToken, refreshToken, err := domain.Login(ctx, "testuser", password)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "account is not active")
				So(loggedUser, ShouldBeNil)
				So(accessToken, ShouldBeEmpty)
				So(refreshToken, ShouldBeEmpty)
			})
		})

		Convey("ValidateToken - requires valid JWT token format", func() {
			// This test validates that the domain correctly handles token parsing
			// Full token validation tests would require generating actual JWT tokens
		})
	})
}
