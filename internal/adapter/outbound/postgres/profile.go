package postgres_outbound_adapter

import (
	"context"
	"database/sql"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/google/uuid"
	"github.com/palantir/stacktrace"

	"prabogo/internal/model"
	outbound_port "prabogo/internal/port/outbound"
	contextutil "prabogo/utils/context"
)

const (
	tableProfiles      = "mikrotik_profiles"
	tableProfilesPPPoE = "mikrotik_profile_pppoe"
)

type profileAdapter struct {
	db outbound_port.DatabaseExecutor
}

func NewProfileAdapter(
	db outbound_port.DatabaseExecutor,
) outbound_port.ProfileDatabasePort {
	return &profileAdapter{
		db: db,
	}
}

func (a *profileAdapter) CreateProfile(ctx context.Context, input model.ProfileInput, mikrotikID uuid.UUID) (*model.Profile, error) {
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get tenant ID from context")
	}

	model.PrepareProfileInput(&input)

	record := goqu.Record{
		"tenant_id":                 tenantID,
		"mikrotik_id":               mikrotikID,
		"name":                      input.Name,
		"profile_type":              "pppoe",
		"mikrotik_object_id":        "", // Will be updated after MikroTik API call
		"rate_limit_up_kbps":        input.RateLimitUpKbps,
		"rate_limit_down_kbps":      input.RateLimitDownKbps,
		"idle_timeout_seconds":      input.IdleTimeoutSeconds,
		"session_timeout_seconds":   input.SessionTimeoutSeconds,
		"keepalive_timeout_seconds": input.KeepaliveTimeoutSeconds,
		"only_one":                  *input.OnlyOne,
		"status_authentication":     *input.StatusAuthentication,
		"dns_server":                input.DNSServer,
		"is_active":                 true,
		"sync_with_mikrotik":        true,
		"price":                     input.Price,
		"tax_rate":                  *input.TaxRate,
	}

	query, _, err := goqu.Dialect("postgres").
		Insert(tableProfiles).
		Rows(record).
		Returning("*").
		ToSQL()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to build insert query")
	}

	var result model.Profile
	err = a.db.QueryRowContext(ctx, query).Scan(
		&result.ID,
		&result.MikrotikID,
		&result.Name,
		&result.ProfileType,
		&result.MikrotikObjectID,
		&result.RateLimitUpKbps,
		&result.RateLimitDownKbps,
		&result.IdleTimeoutSeconds,
		&result.SessionTimeoutSeconds,
		&result.KeepaliveTimeoutSeconds,
		&result.OnlyOne,
		&result.StatusAuthentication,
		&result.DNSServer,
		&result.IsActive,
		&result.SyncWithMikrotik,
		&result.Price,
		&result.TaxRate,
		&result.LastSync,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to insert profile")
	}

	return &result, nil
}

func (a *profileAdapter) CreateProfilePPPoE(ctx context.Context, profileID uuid.UUID, input model.ProfileInput) error {
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "failed to get tenant ID from context")
	}

	model.PrepareProfileInput(&input)

	record := goqu.Record{
		"tenant_id":       tenantID,
		"profile_id":      profileID,
		"local_address":   input.LocalAddress,
		"remote_address":  input.RemoteAddress,
		"address_pool":    input.AddressPool,
		"mtu":             *input.MTU,
		"mru":             *input.MRU,
		"use_mpls":        *input.UseMPLS,
		"use_compression": *input.UseCompression,
		"use_encryption":  *input.UseEncryption,
	}

	query, _, err := goqu.Dialect("postgres").
		Insert(tableProfilesPPPoE).
		Rows(record).
		ToSQL()
	if err != nil {
		return stacktrace.Propagate(err, "failed to build insert pppoe query")
	}

	_, err = a.db.ExecContext(ctx, query)
	if err != nil {
		return stacktrace.Propagate(err, "failed to insert profile pppoe")
	}

	return nil
}

func (a *profileAdapter) UpdateMikrotikObjectID(ctx context.Context, profileID uuid.UUID, objectID string) error {
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "failed to get tenant ID from context")
	}

	query, _, err := goqu.Dialect("postgres").
		Update(tableProfiles).
		Set(goqu.Record{"mikrotik_object_id": objectID}).
		Where(goqu.Ex{"id": profileID, "tenant_id": tenantID}).
		ToSQL()
	if err != nil {
		return stacktrace.Propagate(err, "failed to build update query")
	}

	result, err := a.db.ExecContext(ctx, query)
	if err != nil {
		return stacktrace.Propagate(err, "failed to update mikrotik object id")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return stacktrace.Propagate(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return stacktrace.NewError("profile not found")
	}

	return nil
}

func (a *profileAdapter) GetByID(ctx context.Context, id uuid.UUID) (*model.ProfileWithPPPoE, error) {
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get tenant ID from context")
	}

	// Query profile
	profileQuery, _, err := goqu.Dialect("postgres").
		From(tableProfiles).
		Where(goqu.Ex{"id": id, "tenant_id": tenantID}).
		ToSQL()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to build profile query")
	}

	var profile model.Profile
	err = a.db.QueryRowContext(ctx, profileQuery).Scan(
		&profile.ID,
		&profile.MikrotikID,
		&profile.Name,
		&profile.ProfileType,
		&profile.MikrotikObjectID,
		&profile.RateLimitUpKbps,
		&profile.RateLimitDownKbps,
		&profile.IdleTimeoutSeconds,
		&profile.SessionTimeoutSeconds,
		&profile.KeepaliveTimeoutSeconds,
		&profile.OnlyOne,
		&profile.StatusAuthentication,
		&profile.DNSServer,
		&profile.IsActive,
		&profile.SyncWithMikrotik,
		&profile.Price,
		&profile.TaxRate,
		&profile.LastSync,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, stacktrace.NewError("profile not found")
		}
		return nil, stacktrace.Propagate(err, "failed to get profile")
	}

	// Query PPPoE settings
	pppoeQuery, _, err := goqu.Dialect("postgres").
		From(tableProfilesPPPoE).
		Where(goqu.Ex{"profile_id": id, "tenant_id": tenantID}).
		ToSQL()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to build pppoe query")
	}

	var pppoe model.ProfilePPPoE
	err = a.db.QueryRowContext(ctx, pppoeQuery).Scan(
		&pppoe.ProfileID,
		&pppoe.LocalAddress,
		&pppoe.RemoteAddress,
		&pppoe.AddressPool,
		&pppoe.MTU,
		&pppoe.MRU,
		&pppoe.UseMPLS,
		&pppoe.UseCompression,
		&pppoe.UseEncryption,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, stacktrace.Propagate(err, "failed to get pppoe settings")
	}

	result := &model.ProfileWithPPPoE{
		Profile: profile,
	}

	if err != sql.ErrNoRows {
		result.PPPoE = &pppoe
	}

	return result, nil
}

func (a *profileAdapter) GetByMikrotikID(ctx context.Context, mikrotikID, profileID uuid.UUID) (*model.ProfileWithPPPoE, error) {
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get tenant ID from context")
	}

	// Query profile
	profileQuery, _, err := goqu.Dialect("postgres").
		From(tableProfiles).
		Where(goqu.Ex{"id": profileID, "mikrotik_id": mikrotikID, "tenant_id": tenantID}).
		ToSQL()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to build profile query")
	}

	var profile model.Profile
	err = a.db.QueryRowContext(ctx, profileQuery).Scan(
		&profile.ID,
		&profile.MikrotikID,
		&profile.Name,
		&profile.ProfileType,
		&profile.MikrotikObjectID,
		&profile.RateLimitUpKbps,
		&profile.RateLimitDownKbps,
		&profile.IdleTimeoutSeconds,
		&profile.SessionTimeoutSeconds,
		&profile.KeepaliveTimeoutSeconds,
		&profile.OnlyOne,
		&profile.StatusAuthentication,
		&profile.DNSServer,
		&profile.IsActive,
		&profile.SyncWithMikrotik,
		&profile.Price,
		&profile.TaxRate,
		&profile.LastSync,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, stacktrace.NewError("profile not found")
		}
		return nil, stacktrace.Propagate(err, "failed to get profile")
	}

	// Query PPPoE settings
	pppoeQuery, _, err := goqu.Dialect("postgres").
		From(tableProfilesPPPoE).
		Where(goqu.Ex{"profile_id": profileID, "tenant_id": tenantID}).
		ToSQL()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to build pppoe query")
	}

	var pppoe model.ProfilePPPoE
	err = a.db.QueryRowContext(ctx, pppoeQuery).Scan(
		&pppoe.ProfileID,
		&pppoe.LocalAddress,
		&pppoe.RemoteAddress,
		&pppoe.AddressPool,
		&pppoe.MTU,
		&pppoe.MRU,
		&pppoe.UseMPLS,
		&pppoe.UseCompression,
		&pppoe.UseEncryption,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, stacktrace.Propagate(err, "failed to get pppoe settings")
	}

	result := &model.ProfileWithPPPoE{
		Profile: profile,
	}

	if err != sql.ErrNoRows {
		result.PPPoE = &pppoe
	}

	return result, nil
}

func (a *profileAdapter) List(ctx context.Context, mikrotikID uuid.UUID) ([]model.ProfileWithPPPoE, error) {
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get tenant ID from context")
	}

	// Query all profiles for this MikroTik
	profilesQuery, _, err := goqu.Dialect("postgres").
		From(tableProfiles).
		Where(goqu.Ex{"mikrotik_id": mikrotikID, "tenant_id": tenantID}).
		Order(goqu.I("created_at").Desc()).
		ToSQL()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to build profiles query")
	}

	rows, err := a.db.QueryContext(ctx, profilesQuery)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to query profiles")
	}
	defer rows.Close()

	var profiles []model.Profile
	for rows.Next() {
		var profile model.Profile
		err := rows.Scan(
			&profile.ID,
			&profile.MikrotikID,
			&profile.Name,
			&profile.ProfileType,
			&profile.MikrotikObjectID,
			&profile.RateLimitUpKbps,
			&profile.RateLimitDownKbps,
			&profile.IdleTimeoutSeconds,
			&profile.SessionTimeoutSeconds,
			&profile.KeepaliveTimeoutSeconds,
			&profile.OnlyOne,
			&profile.StatusAuthentication,
			&profile.DNSServer,
			&profile.IsActive,
			&profile.SyncWithMikrotik,
			&profile.Price,
			&profile.TaxRate,
			&profile.LastSync,
			&profile.CreatedAt,
			&profile.UpdatedAt,
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to scan profile")
		}
		profiles = append(profiles, profile)
	}

	// For each profile, get PPPoE settings
	var result []model.ProfileWithPPPoE
	for _, profile := range profiles {
		pppoeQuery, _, err := goqu.Dialect("postgres").
			From(tableProfilesPPPoE).
			Where(goqu.Ex{"profile_id": profile.ID, "tenant_id": tenantID}).
			ToSQL()
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to build pppoe query")
		}

		var pppoe model.ProfilePPPoE
		err = a.db.QueryRowContext(ctx, pppoeQuery).Scan(
			&pppoe.ProfileID,
			&pppoe.LocalAddress,
			&pppoe.RemoteAddress,
			&pppoe.AddressPool,
			&pppoe.MTU,
			&pppoe.MRU,
			&pppoe.UseMPLS,
			&pppoe.UseCompression,
			&pppoe.UseEncryption,
		)

		profileWithPPPoE := model.ProfileWithPPPoE{
			Profile: profile,
		}

		if err != sql.ErrNoRows {
			profileWithPPPoE.PPPoE = &pppoe
		}

		result = append(result, profileWithPPPoE)
	}

	return result, nil
}

func (a *profileAdapter) Update(ctx context.Context, id uuid.UUID, input model.ProfileInput) error {
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "failed to get tenant ID from context")
	}

	model.PrepareProfileInput(&input)

	// Update profile
	profileUpdate := goqu.Record{
		"name":                      input.Name,
		"rate_limit_up_kbps":        input.RateLimitUpKbps,
		"rate_limit_down_kbps":      input.RateLimitDownKbps,
		"idle_timeout_seconds":      input.IdleTimeoutSeconds,
		"session_timeout_seconds":   input.SessionTimeoutSeconds,
		"keepalive_timeout_seconds": input.KeepaliveTimeoutSeconds,
		"only_one":                  *input.OnlyOne,
		"status_authentication":     *input.StatusAuthentication,
		"dns_server":                input.DNSServer,
		"price":                     input.Price,
		"tax_rate":                  *input.TaxRate,
	}

	profileQuery, _, err := goqu.Dialect("postgres").
		Update(tableProfiles).
		Set(profileUpdate).
		Where(goqu.Ex{"id": id, "tenant_id": tenantID}).
		ToSQL()
	if err != nil {
		return stacktrace.Propagate(err, "failed to build profile update query")
	}

	result, err := a.db.ExecContext(ctx, profileQuery)
	if err != nil {
		return stacktrace.Propagate(err, "failed to update profile")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return stacktrace.Propagate(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return stacktrace.NewError("profile not found")
	}

	// Update PPPoE settings
	pppoeUpdate := goqu.Record{
		"local_address":   input.LocalAddress,
		"remote_address":  input.RemoteAddress,
		"address_pool":    input.AddressPool,
		"mtu":             *input.MTU,
		"mru":             *input.MRU,
		"use_mpls":        *input.UseMPLS,
		"use_compression": *input.UseCompression,
		"use_encryption":  *input.UseEncryption,
	}

	pppoeQuery, _, err := goqu.Dialect("postgres").
		Update(tableProfilesPPPoE).
		Set(pppoeUpdate).
		Where(goqu.Ex{"profile_id": id, "tenant_id": tenantID}).
		ToSQL()
	if err != nil {
		return stacktrace.Propagate(err, "failed to build pppoe update query")
	}

	_, err = a.db.ExecContext(ctx, pppoeQuery)
	if err != nil {
		return stacktrace.Propagate(err, "failed to update pppoe settings")
	}

	return nil
}

func (a *profileAdapter) Delete(ctx context.Context, id uuid.UUID) error {
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "failed to get tenant ID from context")
	}

	// Delete profile (PPPoE settings will be cascade deleted)
	query, _, err := goqu.Dialect("postgres").
		Delete(tableProfiles).
		Where(goqu.Ex{"id": id, "tenant_id": tenantID}).
		ToSQL()
	if err != nil {
		return stacktrace.Propagate(err, "failed to build delete query")
	}

	result, err := a.db.ExecContext(ctx, query)
	if err != nil {
		return stacktrace.Propagate(err, "failed to delete profile")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return stacktrace.Propagate(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return stacktrace.NewError("profile not found")
	}

	return nil
}
