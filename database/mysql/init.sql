-- Inicialización de la base de datos MySQL para users-api

-- Crear base de datos si no existe
CREATE DATABASE IF NOT EXISTS gym_users;
USE gym_users;

-- Tabla de usuarios
CREATE TABLE IF NOT EXISTS users (
                                     id INT AUTO_INCREMENT PRIMARY KEY,
                                     username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    first_name VARCHAR(50),
    last_name VARCHAR(50),
    role ENUM('normal', 'admin') DEFAULT 'normal',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_username (username),
    INDEX idx_email (email),
    INDEX idx_role (role)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Insertar usuario admin por defecto (password: admin123)
-- Hash bcrypt de "admin123"
INSERT INTO users (username, email, password, first_name, last_name, role)
VALUES (
           'admin',
           'admin@gym.com',
           '$2a$10$RBKjZOTX/.fYdxV.1DFurObGvOHqOJeTipwLoQ/ZMla8VdHx6QWti',
           'Admin',
           'System',
           'admin'
       ) ON DUPLICATE KEY UPDATE id=id;

-- Insertar usuario normal de prueba (password: user123)
-- Hash bcrypt de "user123"
INSERT INTO users (username, email, password, first_name, last_name, role)
VALUES (
           'testuser',
           'user@gym.com',
           '$2a$10$DlSThreWPj8evRKlPKyRt.Czi0jgHL.FA9GIMvFH93n43gwNbbOdW',
           'Test',
           'User',
           'normal'
       ) ON DUPLICATE KEY UPDATE id=id;