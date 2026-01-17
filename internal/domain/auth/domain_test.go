package auth

import (
	"context"
	"errors"
	"testing"

	"MikrOps/internal/model"
	mock_outbound_port "MikrOps/tests/mocks/port"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/crypto/bcrypt"
)

func TestAuth(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_outbound_port.NewMockDatabasePort(ctrl)
	mockAuthDB := mock_outbound_port.NewMockAuthDatabasePort(ctrl)

	mockDB.EXPECT().Auth().Return(mockAuthDB).AnyTimes()

	domain := NewAuthDomain(mockDB)
	ctx := context.Background()

	Convey("Test Auth Domain", t, func() {
		Convey("Register", func() {
			input := model.UserInput{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
				Fullname: "Test User",
			}

			Convey("Email already registered", func() {
				mockAuthDB.EXPECT().FindUserByFilter(gomock.Any(), false).Return([]model.User{{ID: uuid.New()}}, nil)
				user, err := domain.Register(ctx, input)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "email already registered")
				So(user, ShouldBeNil)
			})

			Convey("Success with default role", func() {
				mockAuthDB.EXPECT().FindUserByFilter(gomock.Any(), false).Return([]model.User{}, nil)
				mockAuthDB.EXPECT().FindRoleByName("viewer").Return(&model.Role{ID: uuid.New(), Name: "viewer"}, nil)
				mockAuthDB.EXPECT().SaveUser(gomock.Any()).Return(nil)

				user, err := domain.Register(ctx, input)
				So(err, ShouldBeNil)
				So(user, ShouldNotBeNil)
				So(user.Email, ShouldEqual, input.Email)
				So(user.UserRole, ShouldEqual, model.RoleViewer)
			})

			Convey("Database error", func() {
				mockAuthDB.EXPECT().FindUserByFilter(gomock.Any(), false).Return(nil, errors.New("db error"))
				user, err := domain.Register(ctx, input)
				So(err, ShouldNotBeNil)
				So(user, ShouldBeNil)
			})
		})

		Convey("Login", func() {
			password := "password123"
			hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			user := model.User{
				ID:                uuid.New(),
				Username:          "testuser",
				Email:             "test@example.com",
				EncryptedPassword: string(hashed),
				Status:            model.UserStatusActive,
				UserRole:          model.RoleAdmin,
			}

			Convey("Success", func() {
				mockAuthDB.EXPECT().FindUserByFilter(gomock.Any(), false).Return([]model.User{user}, nil)
				loggedUser, token, err := domain.Login(ctx, "testuser", password)
				So(err, ShouldBeNil)
				So(loggedUser, ShouldNotBeNil)
				So(loggedUser.ID, ShouldEqual, user.ID)
				So(token, ShouldNotBeEmpty)
			})

			Convey("Invalid credentials - user not found", func() {
				mockAuthDB.EXPECT().FindUserByFilter(gomock.Any(), false).Return([]model.User{}, nil)
				loggedUser, token, err := domain.Login(ctx, "wrong", password)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "invalid credentials")
				So(loggedUser, ShouldBeNil)
				So(token, ShouldBeEmpty)
			})

			Convey("Invalid credentials - wrong password", func() {
				mockAuthDB.EXPECT().FindUserByFilter(gomock.Any(), false).Return([]model.User{user}, nil)
				loggedUser, token, err := domain.Login(ctx, "testuser", "wrong")
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "invalid credentials")
				So(loggedUser, ShouldBeNil)
				So(token, ShouldBeEmpty)
			})

			Convey("Account inactive", func() {
				inactiveUser := user
				inactiveUser.Status = model.UserStatusInactive
				mockAuthDB.EXPECT().FindUserByFilter(gomock.Any(), false).Return([]model.User{inactiveUser}, nil)
				loggedUser, token, err := domain.Login(ctx, "testuser", password)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "account is not active")
				So(loggedUser, ShouldBeNil)
				So(token, ShouldBeEmpty)
			})
		})

		Convey("ValidateToken", func() {
			// First login to get a valid token
			password := "password123"
			hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			user := model.User{
				ID:                uuid.New(),
				Username:          "testuser",
				Email:             "test@example.com",
				EncryptedPassword: string(hashed),
				Status:            model.UserStatusActive,
				UserRole:          model.RoleAdmin,
			}

			mockAuthDB.EXPECT().FindUserByFilter(gomock.Any(), false).Return([]model.User{user}, nil)
			_, token, _ := domain.Login(ctx, "testuser", password)

			Convey("Success", func() {
				mockAuthDB.EXPECT().FindUserByFilter(gomock.Any(), false).Return([]model.User{user}, nil)
				validatedUser, err := domain.ValidateToken(ctx, token)
				So(err, ShouldBeNil)
				So(validatedUser, ShouldNotBeNil)
				So(validatedUser.ID, ShouldEqual, user.ID)
			})

			Convey("Invalid token", func() {
				validatedUser, err := domain.ValidateToken(ctx, "invalid-token")
				So(err, ShouldNotBeNil)
				So(validatedUser, ShouldBeNil)
			})

			Convey("User not found in DB during validation", func() {
				mockAuthDB.EXPECT().FindUserByFilter(gomock.Any(), false).Return([]model.User{}, nil)
				validatedUser, err := domain.ValidateToken(ctx, token)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "user not found")
				So(validatedUser, ShouldBeNil)
			})
		})
	})
}

