-- Удаляем старую таблицу, если она существует
DROP TABLE IF EXISTS users;

-- Создаем новую таблицу
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    ip TEXT NOT NULL,
    duration BIGINT,
    success_date TIMESTAMP
);