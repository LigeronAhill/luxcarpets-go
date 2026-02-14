-- +tern:Up
-- Создаем тип ENUM для ролей
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_role') THEN
        CREATE TYPE user_role AS ENUM ('guest', 'customer', 'employee', 'admin', 'owner');
    END IF;
END$$;

-- Создаем таблицу пользователей
CREATE TABLE IF NOT EXISTS users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email VARCHAR(255) NOT NULL UNIQUE,
  email_verified BOOLEAN NOT NULL DEFAULT FALSE,
  verification_token VARCHAR(255),
  username VARCHAR(255) NOT NULL,
  role user_role NOT NULL DEFAULT 'guest',
  image_url VARCHAR(255),
  password_hash VARCHAR(255),
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMP,
  -- Дополнительные ограничения
  CONSTRAINT chk_email CHECK (
    email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'
  ),
  CONSTRAINT chk_username_length CHECK (LENGTH(username) BETWEEN 3 AND 50)
);

-- Создаем индексы
CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);

CREATE INDEX IF NOT EXISTS idx_users_username ON users (username);

CREATE INDEX IF NOT EXISTS idx_users_role ON users (role);

CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users (deleted_at)
WHERE
  deleted_at IS NULL;

-- Функция для обновления поля updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column () RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Создаем триггер для автоматического обновления updated_at
CREATE OR REPLACE TRIGGER update_users_updated_at BEFORE
UPDATE ON users FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column ();

-- Комментарии
COMMENT ON TABLE users IS 'Пользователи системы';

COMMENT ON COLUMN users.email IS 'Электронная почта (уникальный идентификатор)';

COMMENT ON COLUMN users.password_hash IS 'Хэш пароля';

COMMENT ON COLUMN users.role IS 'Роль пользователя: guest, customer, employee, admin, owner';

COMMENT ON COLUMN users.deleted_at IS 'Мягкое удаление - если NULL, то пользователь активен';

COMMENT ON FUNCTION update_updated_at_column () IS 'Функция для автоматического обновления поля updated_at';

COMMENT ON TRIGGER update_users_updated_at ON users IS 'Триггер для автоматического обновления updated_at при изменении записи';

---- create above / drop below ----
-- Удаляем триггер
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Удаляем функцию
DROP FUNCTION IF EXISTS update_updated_at_column;

-- Удаляем индексы
DROP INDEX IF EXISTS idx_users_deleted_at;

DROP INDEX IF EXISTS idx_users_role;

DROP INDEX IF EXISTS idx_users_username;

DROP INDEX IF EXISTS idx_users_email;

-- Удаляем таблицу
DROP TABLE IF EXISTS users CASCADE;

-- Удаляем тип ENUM
DROP TYPE IF EXISTS user_role;
