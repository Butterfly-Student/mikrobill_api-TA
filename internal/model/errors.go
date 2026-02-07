package model

import "errors"

var (
	// ErrCacheNotFound is returned when a cache lookup fails to find a value
	ErrCacheNotFound = errors.New("cache entry not found")

	// ErrTokenExpired is returned when a JWT token has expired
	ErrTokenExpired = errors.New("token has expired")

	// ErrTokenInvalid is returned when a JWT token is invalid
	ErrTokenInvalid = errors.New("token is invalid")

	// ErrUserNotFound is returned when a user is not found
	ErrUserNotFound = errors.New("user not found")

	// ErrTenantNotFound is returned when a tenant is not found
	ErrTenantNotFound = errors.New("tenant not found")

	// ErrUnauthorized is returned when authentication fails
	ErrUnauthorized = errors.New("unauthorized access")

	// ErrForbidden is returned when access is denied
	ErrForbidden = errors.New("forbidden access")

	// ErrInvalidCredentials is returned when credentials are invalid
	ErrInvalidCredentials = errors.New("invalid credentials")

	// ErrInvalidInput is returned when input validation fails
	ErrInvalidInput = errors.New("invalid input")

	// ErrResourceNotFound is returned when a resource is not found
	ErrResourceNotFound = errors.New("resource not found")

	// ErrDuplicateResource is returned when a duplicate resource is attempted
	ErrDuplicateResource = errors.New("resource already exists")
)
