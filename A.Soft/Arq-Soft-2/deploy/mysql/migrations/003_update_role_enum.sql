-- Primero permitimos ambos valores en el enum
ALTER TABLE users MODIFY COLUMN role ENUM('user', 'normal', 'admin') NOT NULL DEFAULT 'user';

-- Actualizamos los valores antiguos
UPDATE users SET role = 'user' WHERE role = 'normal';