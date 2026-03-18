-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Таблица пользователей
CREATE TABLE users (
    user_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) NOT NULL UNIQUE,
    phone VARCHAR(20) UNIQUE,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    subscription_status VARCHAR(20) NOT NULL DEFAULT 'NONE',
    subscription_expires TIMESTAMPTZ,
    
    CONSTRAINT check_subscription_status CHECK (subscription_status IN ('ACTIVE', 'INACTIVE', 'EXPIRED', 'TRIAL', 'NONE')),
    CONSTRAINT check_email_format CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$')
);

-- Индексы для users
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_phone ON users(phone);
CREATE INDEX idx_users_subscription_status ON users(subscription_status);
CREATE INDEX idx_users_is_active ON users(is_active);

-- Таблица профилей пользователей
CREATE TABLE user_profiles (
    profile_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    surname VARCHAR(100) NOT NULL,
    patronymic VARCHAR(100),
    date_of_birth DATE,
    gender VARCHAR(10),
    height_cm INTEGER,
    weight_kg DECIMAL(5,2),
    fitness_goal VARCHAR(20),
    experience_level VARCHAR(20),
    health_limitations TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT fk_user_profiles_user FOREIGN KEY (user_id) 
        REFERENCES users(user_id) ON DELETE CASCADE,
    CONSTRAINT check_gender CHECK (gender IN ('MALE', 'FEMALE', 'OTHER')),
    CONSTRAINT check_fitness_goal CHECK (fitness_goal IN ('WEIGHT_LOSS', 'MUSCLE_GAIN', 'ENDURANCE', 'GENERAL_HEALTH', 'MAINTENANCE')),
    CONSTRAINT check_experience_level CHECK (experience_level IN ('BEGINNER', 'INTERMEDIATE', 'ADVANCED', 'PROFESSIONAL')),
    CONSTRAINT check_height_positive CHECK (height_cm > 0),
    CONSTRAINT check_weight_positive CHECK (weight_kg > 0)
);

-- Индексы для user_profiles
CREATE INDEX idx_user_profiles_user_id ON user_profiles(user_id);
CREATE INDEX idx_user_profiles_name ON user_profiles(name, surname);
CREATE INDEX idx_user_profiles_fitness_goal ON user_profiles(fitness_goal);
CREATE INDEX idx_user_profiles_experience_level ON user_profiles(experience_level);

-- Таблица настроек пользователей
CREATE TABLE user_settings (
    settings_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL UNIQUE,
    notifications_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    language VARCHAR(10) NOT NULL DEFAULT 'en',
    timezone VARCHAR(50) NOT NULL DEFAULT 'UTC',
    privacy_level VARCHAR(10) NOT NULL DEFAULT 'PRIVATE',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT fk_user_settings_user FOREIGN KEY (user_id) 
        REFERENCES users(user_id) ON DELETE CASCADE,
    CONSTRAINT check_privacy_level CHECK (privacy_level IN ('PUBLIC', 'PRIVATE', 'FRIENDS', 'ONLY_ME')),
    CONSTRAINT check_language_format CHECK (language ~ '^[a-z]{2}(-[A-Z]{2})?$')
);

-- Индексы для user_settings
CREATE INDEX idx_user_settings_user_id ON user_settings(user_id);
CREATE INDEX idx_user_settings_privacy_level ON user_settings(privacy_level);

-- Функция для автоматического обновления updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Триггеры для автоматического обновления updated_at
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_user_profiles_updated_at
    BEFORE UPDATE ON user_profiles
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_user_settings_updated_at
    BEFORE UPDATE ON user_settings
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Комментарии к таблицам и колонкам
COMMENT ON TABLE users IS 'Основная информация о пользователях';
COMMENT ON COLUMN users.user_id IS 'Уникальный идентификатор пользователя';
COMMENT ON COLUMN users.email IS 'Email пользователя';
COMMENT ON COLUMN users.phone IS 'Номер телефона';
COMMENT ON COLUMN users.password IS 'Хеш пароля';
COMMENT ON COLUMN users.subscription_status IS 'Статус подписки';
COMMENT ON COLUMN users.subscription_expires IS 'Дата истечения подписки';

COMMENT ON TABLE user_profiles IS 'Профили пользователей с дополнительной информацией';
COMMENT ON TABLE user_settings IS 'Настройки пользователей';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP TRIGGER IF EXISTS update_user_profiles_updated_at ON user_profiles;
DROP TRIGGER IF EXISTS update_user_settings_updated_at ON user_settings;

DROP FUNCTION IF EXISTS update_updated_at_column();

DROP TABLE IF EXISTS user_settings;
DROP TABLE IF EXISTS user_profiles;
DROP TABLE IF EXISTS users;

DROP EXTENSION IF EXISTS "uuid-ossp";
-- +goose StatementEnd