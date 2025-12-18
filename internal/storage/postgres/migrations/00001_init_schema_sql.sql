-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
  id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
  login VARCHAR UNIQUE NOT NULL,
  password_hash VARCHAR NOT NULL,
  encrypted_secret VARCHAR NOT NULL
);
COMMENT ON COLUMN users.login IS 'Логин пользователя';
COMMENT ON COLUMN users.password_hash IS 'Хэш пароля пользователя';
COMMENT ON COLUMN users.encrypted_secret IS 'Зашифрованный ключ пользователя (для данных)';

CREATE TABLE user_data (
  id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users (id),
  encrypted_data BYTEA NOT NULL,
  meta JSONB NOT NULL,
  data_type VARCHAR NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()  
);
COMMENT ON COLUMN user_data.encrypted_data IS 'Зашифрованные данные';
COMMENT ON COLUMN user_data.meta IS 'Мета информация о данных';
COMMENT ON COLUMN user_data.data_type IS 'Тип данных';
COMMENT ON COLUMN user_data.created_at IS 'Timestamp создания записи';
COMMENT ON COLUMN user_data.updated_at IS 'Timestamp обновления записи';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE users;
DROP TABLE user_data;
-- +goose StatementEnd