-- +goose Down
-- +goose StatementBegin

DROP TRIGGER IF EXISTS ensure_profile_child_exists ON mikrotik_profiles;
DROP TRIGGER IF EXISTS validate_package_profile_trigger ON packages;
DROP FUNCTION IF EXISTS validate_profile_child_exists();
DROP FUNCTION IF EXISTS validate_package_profile();

DROP TRIGGER IF EXISTS set_updated_at_packages ON packages;
DROP TRIGGER IF EXISTS set_updated_at_mikrotik_profiles ON mikrotik_profiles;

DROP TABLE IF EXISTS packages;
DROP TABLE IF EXISTS mikrotik_queue_settings;
DROP TABLE IF EXISTS mikrotik_profile_static_ip;
DROP TABLE IF EXISTS mikrotik_profile_hotspot;
DROP TABLE IF EXISTS mikrotik_profile_pppoe;
DROP TABLE IF EXISTS mikrotik_profiles;

DROP TYPE IF EXISTS profile_type;

-- +goose StatementEnd