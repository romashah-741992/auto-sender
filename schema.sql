CREATE DATABASE IF NOT EXISTS auto_sender CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE auto_sender;

CREATE TABLE IF NOT EXISTS messages (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    recipient VARCHAR(32) NOT NULL,
    content VARCHAR(500) NOT NULL,
    status ENUM('pending','sent','failed') NOT NULL DEFAULT 'pending',
    external_message_id VARCHAR(128) NULL,
    error_text TEXT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    sent_at TIMESTAMP NULL,
    INDEX idx_status_created_at (status, created_at)
);

INSERT INTO messages (recipient, content, status)
VALUES
('+905551111111', 'Insider - Project 1', 'pending'),
('+905552222222', 'Insider - Project 2', 'pending'),
('+905553333333', 'Another test message', 'pending'),
('+905553333999', 'Another test message2', 'pending'),
('+905553333977', 'Another test message3 Another test message3 Another test message4', 'pending');

