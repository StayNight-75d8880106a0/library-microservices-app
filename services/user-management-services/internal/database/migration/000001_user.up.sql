CREATE TYPE profile_status_enum AS ENUM ('INCOMPLETE', 'ACTIVE', 'SUSPENDED');

DROP TABLE IF EXISTS "user_profiles" CASCADE;
CREATE TABLE "user_profiles" (
    keycloak_user_id UUID PRIMARY KEY NOT NULL UNIQUE,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    phone_number VARCHAR(20),
    address TEXT,
    profile_status profile_status_enum DEFAULT 'INCOMPLETE',
    library_card_number VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_user_profiles_keycloak_id ON user_profiles(keycloak_user_id);
CREATE INDEX idx_user_profiles_profile_status ON user_profiles(profile_status);
CREATE INDEX idx_user_profiles_name ON user_profiles(first_name, last_name);