package outbound_port

//go:generate mockgen -source=password_hasher_port.go -destination=./../../../tests/mocks/port/mock_password_hasher_port.go

type PasswordHasherPort interface {
	HashPassword(password string) (string, error)
	VerifyPassword(hashedPassword, password string) (bool, error)
	GenerateRandomToken(length int) string
}
