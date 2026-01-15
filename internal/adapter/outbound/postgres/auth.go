package postgres_outbound_adapter

import (
	"context"
	"database/sql"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/google/uuid"
	"github.com/palantir/stacktrace"

	"prabogo/internal/model"
	outbound_port "prabogo/internal/port/outbound"
)

const (
	tableUsers = "users"
	tableRoles = "roles"
)

type authAdapter struct {
	db outbound_port.DatabaseExecutor
}

func NewAuthAdapter(db outbound_port.DatabaseExecutor) outbound_port.AuthDatabasePort {
	return &authAdapter{db: db}
}

func (a *authAdapter) SaveUser(ctx context.Context, user model.User) error {
	record := goqu.Record{
		"id":                    user.ID,
		"tenant_id":             user.TenantID,
		"username":              user.Username,
		"email":                 user.Email,
		"encrypted_password":    user.EncryptedPassword,
		"fullname":              user.Fullname,
		"phone":                 user.Phone,
		"avatar":                user.Avatar,
		"role_id":               user.RoleID,
		"user_role":             user.UserRole,
		"status":                user.Status,
		"is_superadmin":         user.IsSuperadmin,
		"last_login_at":         user.LastLoginAt,
		"last_ip":               user.LastIP,
		"failed_login_attempts": user.FailedLoginAttempts,
		"locked_until":          user.LockedUntil,
		"password_changed_at":   user.PasswordChangedAt,
		"force_password_change": user.ForcePasswordChange,
		"two_factor_enabled":    user.TwoFactorEnabled,
		"two_factor_secret":     user.TwoFactorSecret,
		"created_by":            user.CreatedBy,
		"updated_by":            user.UpdatedBy,
		"created_at":            user.CreatedAt,
		"updated_at":            user.UpdatedAt,
		"deleted_at":            user.DeletedAt,
	}

	query, _, err := goqu.Dialect("postgres").
		Insert(tableUsers).
		Rows(record).
		OnConflict(goqu.DoUpdate("id", record)).
		ToSQL()

	if err != nil {
		return stacktrace.Propagate(err, "failed to build save user query")
	}

	_, err = a.db.ExecContext(ctx, query)
	if err != nil {
		return stacktrace.Propagate(err, "failed to execute save user query")
	}

	return nil
}

func (a *authAdapter) FindUserByFilter(ctx context.Context, filter model.UserFilter, lock bool) ([]model.User, error) {
	ds := goqu.Dialect("postgres").From(tableUsers)

	if len(filter.IDs) > 0 {
		ds = ds.Where(goqu.Ex{"id": filter.IDs})
	}

	var orConditions []goqu.Expression
	if len(filter.Usernames) > 0 {
		orConditions = append(orConditions, goqu.Ex{"username": filter.Usernames})
	}
	if len(filter.Emails) > 0 {
		orConditions = append(orConditions, goqu.Ex{"email": filter.Emails})
	}

	if len(orConditions) > 0 {
		ds = ds.Where(goqu.Or(orConditions...))
	}

	if lock {
		ds = ds.ForUpdate(exp.Wait)
	}

	query, _, err := ds.ToSQL()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to build find users query")
	}

	rows, err := a.db.QueryContext(ctx, query)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to execute find users query")
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var user model.User
		err := rows.Scan(
			&user.ID,
			&user.TenantID,
			&user.Username,
			&user.Email,
			&user.EncryptedPassword,
			&user.Fullname,
			&user.Phone,
			&user.Avatar,
			&user.RoleID,
			&user.UserRole,
			&user.Status,
			&user.IsSuperadmin,
			&user.LastLoginAt,
			&user.LastIP,
			&user.FailedLoginAttempts,
			&user.LockedUntil,
			&user.PasswordChangedAt,
			&user.ForcePasswordChange,
			&user.TwoFactorEnabled,
			&user.TwoFactorSecret,
			&user.CreatedBy,
			&user.UpdatedBy,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.DeletedAt,
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to scan user row")
		}
		users = append(users, user)
	}

	return users, nil
}

func (a *authAdapter) FindRoleByName(ctx context.Context, name string) (*model.Role, error) {
	query, _, err := goqu.Dialect("postgres").
		From(tableRoles).
		Where(goqu.Ex{"name": name}).
		Limit(1).
		ToSQL()

	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to build find role by name query")
	}

	var role model.Role
	err = a.db.QueryRowContext(ctx, query).Scan(
		&role.ID,
		&role.TenantID,
		&role.Name,
		&role.DisplayName,
		&role.Description,
		&role.Permissions,
		&role.IsSystem,
		&role.IsActive,
		&role.CreatedAt,
		&role.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, stacktrace.Propagate(err, "failed to execute find role by name query")
	}

	return &role, nil
}

func (a *authAdapter) FindRoleByID(ctx context.Context, id uuid.UUID) (*model.Role, error) {
	query, _, err := goqu.Dialect("postgres").
		From(tableRoles).
		Where(goqu.Ex{"id": id}).
		Limit(1).
		ToSQL()

	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to build find role by id query")
	}

	var role model.Role
	err = a.db.QueryRowContext(ctx, query).Scan(
		&role.ID,
		&role.TenantID,
		&role.Name,
		&role.DisplayName,
		&role.Description,
		&role.Permissions,
		&role.IsSystem,
		&role.IsActive,
		&role.CreatedAt,
		&role.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, stacktrace.Propagate(err, "failed to execute find role by id query")
	}

	return &role, nil
}
