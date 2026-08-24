DROP INDEX IF EXISTS idx_user_profiles_profile_status;
DROP INDEX IF EXISTS idx_user_profiles_name;
DROP INDEX IF EXISTS idx_user_profiles_keycloak_id;

DROP TABLE IF EXISTS "user_profiles" CASCADE;
DROP TYPE IF EXISTS profile_status_enum;