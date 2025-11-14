DROP TABLE IF EXISTS users_temp;

-- Crear una tabla temporal con la nueva estructura
CREATE TABLE users_temp (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(50) NOT NULL,
    email VARCHAR(120) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role ENUM('user', 'admin') NOT NULL DEFAULT 'user',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY username_idx (username),
    UNIQUE KEY email_idx (email)
);

-- Insertar los datos existentes, convirtiendo 'normal' a 'user'
INSERT INTO users_temp (id, username, email, password_hash, role, created_at, updated_at)
SELECT 
    id, 
    username, 
    email, 
    password_hash,
    CASE 
        WHEN role = 'normal' THEN 'user'
        ELSE role
    END as role,
    created_at,
    updated_at
FROM users;

-- Eliminar la tabla original
DROP TABLE users;

-- Renombrar la tabla temporal
RENAME TABLE users_temp TO users;