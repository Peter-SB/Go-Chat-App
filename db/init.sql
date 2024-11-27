CREATE DATABASE IF NOT EXISTS chatapp;

USE chatapp;

-- Messages table
CREATE TABLE IF NOT EXISTS messages (
    id INT AUTO_INCREMENT PRIMARY KEY,
    sender VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    timestamp DATETIME NOT NULL
);

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id INT AUTO_INCREMENT PRIMARY KEY,                              -- Unique identifier for each user
    username VARCHAR(255) NOT NULL UNIQUE,                          -- Username (must be unique)
    hashed_password VARCHAR(255) NOT NULL,                          -- Password hash
    session_token VARCHAR(255) NOT NULL DEFAULT '',                 -- Session token for authentication
    csrf_token VARCHAR(255) NOT NULL DEFAULT '',                    -- CSRF token for request validation
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,                  -- Account creation timestamp
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP -- Last update timestamp
);